package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"

	"github.com/peterbourgon/ff/v3/ffcli"
)

const notionBaseAddress = "https://api.notion.com/v1"

type Config struct {
    DBID string
    NotionKey string
    HttpClient httpInteractor
    Cmd commandOutputter
    DateForEntries string
}

type commandOutputter interface {
    Output() ([]byte, error)
}

type httpInteractor interface {
    Do(*http.Request) (*http.Response, error)
}

type JrnlBody struct {
    Entries []Entry `json:"entries"`
}

type Entry struct {
    Body string `json:"body"`
    Date string `json:"date"`
}

var ErrJrnlCommandFailed = errors.New("the command to get output from jrnl failed with: ")
var ErrFailedToUnmarshalJrnlOutput = errors.New("failed to unmarshal jrnl output: ")

var ErrPostingToNotion = errors.New("internal error making request to notion: ")
var ErrHTTPStatus = errors.New("posting to notion failed with status code: ")

func NewNotionSyncFlagSet(httpClient httpInteractor, cmd commandOutputter, dateForentries string) *ffcli.Command {
    c := &Config{HttpClient: httpClient, Cmd: cmd, DateForEntries: dateForentries}
    syncFlagSet := flag.NewFlagSet("jrnlsync notion", flag.ExitOnError)
    syncFlagSet.StringVar(&c.DBID, "d", "", "The id of the notion database to put the daily journal page")
    syncFlagSet.StringVar(&c.NotionKey, "k", "", "Your notion integration key")


    return &ffcli.Command{
        Name:       "sync",
        ShortUsage: "jrnlSync notion -d [DATABASE_ID] -k [NOTION_INTEGRATION_KEY]",
        ShortHelp:  "Syncs notes from yesterday to your notion database for backup",
        FlagSet:    syncFlagSet,
        Exec:       c.Exec,
    }
}

func (c *Config) Exec(_ context.Context, _ []string) error {
    entriesGroupedByDate, err := c.getEntriesGroupedByDate()
    if err != nil {
        return err
    }
    notionDocument := newNotionDocument(entriesGroupedByDate[c.DateForEntries], c)
    err = c.postToNotion(notionDocument)
    if err != nil {
        return err
    }
    return nil
}

func (c *Config) getEntriesGroupedByDate() (map[string][]string, error) {
    jrnlOutput, err := c.Cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("%w%s", ErrJrnlCommandFailed, err)
    }
    jrnlResp := JrnlBody{}
    err = json.Unmarshal(jrnlOutput, &jrnlResp)
    if err != nil {
        return nil, fmt.Errorf("%w%s", ErrFailedToUnmarshalJrnlOutput, err)
    }
    groupByDate := make(map[string][]string)
    for _, e := range jrnlResp.Entries {
        groupByDate[e.Date] = append(groupByDate[e.Date], e.Body)
    }

    return groupByDate, nil
}

func (c *Config) postToNotion(notionDocument NotionDocument) error {
    jsonBytes, err := json.Marshal(notionDocument)
    if err != nil {
        return err
    }
    buf := bytes.NewBuffer(jsonBytes)

    url := fmt.Sprintf("%s/pages", notionBaseAddress)
    req, err := http.NewRequest("POST", url, buf)
    if err != nil {
        return err
    }

    req.Header = http.Header{
        "Content-Type": []string{"application/json"},
        "Authorization": []string{fmt.Sprintf("Bearer %s", c.NotionKey)},
        "Notion-Version": []string{"2021-08-16"},
    }

    res, err := c.HttpClient.Do(req)
    if err != nil {
        return fmt.Errorf("%w%s", ErrPostingToNotion, err)
    }

    if res.StatusCode > 299 {
        return fmt.Errorf("%w%s", ErrHTTPStatus, res.Status)
    }
    return nil

}
