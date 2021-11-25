package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"time"

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
    Do(*http.Request) (*http.Response, error)
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
    Children []BulletedListItem `json:"children"`
}

type ParentInfo struct {
    DatabaseID string `json:"database_id"`
}

type NotionProperties struct {
    Name NotionName `json:"Name"`
}

type NotionName struct {
    Title []NotionTitle `json:"title"`
}

type NotionTitle struct {
    Text map[string]string `json:"text"`
    Type *string `json:"type,omitempty"`
}


type BulletedListItem struct {
    Object string `json:"object"`
    Type string `json:"type"`
    BulletedList ListItem `json:"bulleted_list_item"`
}

type ListItem struct {
    Text []NotionTitle `json:"text"`
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
    entriesGroupedByDate, err := c.getEntriesGroupedByDate()
    if err != nil {
        return err
    }
    notionDocument := c.buildNotionDocument(entriesGroupedByDate)
    err = c.postToNotion(notionDocument)
    if err != nil {
        return err
    }
    return nil
}

func (c *Config) getEntriesGroupedByDate() (map[string][]string, error) {
    c.cmd.Stdout = c.out
    err := c.cmd.Run()
    if err != nil {
        return nil, err
    }
    jrnlResp := JrnlBody{}
    err = json.Unmarshal(c.out.Bytes(), &jrnlResp)
    if err != nil {
        return nil, err
    }
    groupByDate := make(map[string][]string)
    fmt.Println(len(jrnlResp.Entries))
    for _, e := range jrnlResp.Entries {
        groupByDate[e.Date] = append(groupByDate[e.Date], e.Body)
    }

    return groupByDate, nil
}

func (c *Config) buildNotionDocument(entriesGroupedByDate map[string][]string) NotionDocument {
    children := make([]BulletedListItem, 0)

    time := time.Now().Format("2006-01-02")
    txt := "text"
    todaysEntries := entriesGroupedByDate[time]
    for _, e := range todaysEntries {
        item := BulletedListItem{
                Object: "block",
                Type: "bulleted_list_item",
                BulletedList: ListItem{
                    Text: []NotionTitle{
                        {
                            Type: &txt,
                            Text: map[string]string{"content": e},
                        },
                    },
                },
            }
        children = append(children, item)
    }

    n := NotionDocument{
        Parent: ParentInfo{DatabaseID: c.DBID},
        Properties: NotionProperties{
            Name: NotionName{
                Title: []NotionTitle{
                    {
                        Text: map[string]string{
                            "content": time,
                        },
                    },
                },
            },
        },
        Children: children,
    }
    return n
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

    res, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }

    if res.StatusCode > 299 {
        return fmt.Errorf("posting to notion failed with %d", res.StatusCode)
    }
    return nil

}
