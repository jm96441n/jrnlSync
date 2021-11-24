package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os/exec"

	"github.com/peterbourgon/ff/v3/ffcli"
)

const notionBaseAddress = "https://api.notion.com/v1"

type Config struct {
    DBID string
    NotionKey string
    httpClient httpInteractor
    cmd *exec.Cmd
    out *bytes.Buffer
}

type httpInteractor interface {
    Get(string) (*http.Response, error)
    Post(string, string, io.Reader) (*http.Response, error)
}

type JrnlBody struct {
    Entries []Entry `json:"entries"`
}

type Entry struct {
    Body string `json:"body"`
    Date string `json:"date"`
}

type NotionDocument struct {
    Parent ParentInfo `json:"parent"`
    Properties NotionProperties `json:"properties"`
    Children []ChildInfo `json:"children"`
}

type ParentInfo struct {
    DatabaseID string `json:"database_id"`
}

type NotionProperties struct {
    Name NotionName `json:"Name"`
}

type NotionName struct {
    Title []NotionText `json:"title"`
}

type NotionTitle struct {
    Text NotionText `json:"text"`
}

type NotionText struct {
    Data map[string]string
}

type ChildInfo struct {
    Blocks []BulletedListItems
}

func NewSyncFlagSet(httpClient httpInteractor, output *bytes.Buffer, cmd *exec.Cmd) *ffcli.Command {
    c := &Config{httpClient: httpClient, out: output, cmd: cmd}
    syncFlagSet := flag.NewFlagSet("jrnlNotion sync", flag.ExitOnError)
    syncFlagSet.StringVar(&c.DBID, "d", "", "The id of the notion database to put the daily journal page")
    syncFlagSet.StringVar(&c.NotionKey, "k", "", "Your notion integration key")


    return &ffcli.Command{
        Name:       "sync",
        ShortUsage: "jrnlNotion sync -d [DATABASE_ID] -k [NOTION_INTEGRATION_KEY]",
        ShortHelp:  "Syncs notes from yesterday to your notion database for backup",
        FlagSet:    syncFlagSet,
        Exec:       c.Exec,
    }
}

func (c *Config) Exec(_ context.Context, _ []string) error {
    c.cmd.Stdout = c.out
    err := c.cmd.Run()
    if err != nil {
        return err
    }
    jrnlResp := JrnlBody{}
    err = json.Unmarshal(c.out.Bytes(), &jrnlResp)
    if err != nil {
        return err
    }
    groupByDate := make(map[string][]string)
    for _, e := range jrnlResp.Entries {
        groupByDate[e.Date] = append(groupByDate[e.Date], e.Body)
    }
}
