package sync_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/jm96441n/jrnlNotion/sync"
)

func TestExecSendsResultsFromJrnlToNotionForOnlyTheDateProvided(t *testing.T) {
    tooEarly := map[string]string{
        "body": "Too early",
        "date": "2021-11-23",
    }
    rightDay := []map[string]string{
        {
          "body": "The new one",
          "date": "2021-11-24",
        },
        {
          "body": "next one",
          "date": "2021-11-24",
        },
        {
          "body": "the last one",
          "date": "2021-11-24",
        },
    }
    tooLate := map[string]string{
        "body": "Too late",
        "date": "2021-11-25",
    }

    outputString, err := buildOutputString(tooEarly, rightDay, tooLate)
    if err != nil {
        t.Error(err)
    }

    httpClient := &mockHTTPClient{errOnDo: false}
    config := sync.Config{
        DBID: "mockdbid",
        NotionKey: "fakeNotionKey",
        HttpClient: httpClient,
        Cmd: mockCommand{errOnOutput: false, outputString: outputString},
        DateForEntries: "2021-11-24",
    }
    config.Exec(context.Background(), []string{})
    sentNotionDocument := sync.NotionDocument{}
    err = json.Unmarshal(httpClient.bodyOfRequest, &sentNotionDocument)
    if err != nil {
        t.Error(err)
    }
    notionTitleOfDocument := sentNotionDocument.Properties.Name.Title[0].Text["content"]
    if notionTitleOfDocument != config.DateForEntries {
        t.Errorf("expected %s for the document title, got %s", config.DateForEntries, notionTitleOfDocument)
    }

    if len(sentNotionDocument.Children) != len(rightDay) {
        t.Errorf("expected %d children, got %d", len(rightDay), len(sentNotionDocument.Children))
    }
}

func buildOutputString(tooEarly map[string]string, rightDay []map[string]string, tooLate map[string]string) (string, error) {
    tooEarlyJson, err := json.Marshal(tooEarly)
    if err != nil {
        return "", err
    }
    rightDayString := ""
    for _, m := range rightDay {
        rightDayJson, err := json.Marshal(m)
        if err != nil {
            return "", err
        }
        rightDayString = fmt.Sprintf("%s, %s", rightDayJson, rightDayString)
    }
    tooLateJson, err := json.Marshal(tooLate)
    if err != nil {
        return "", err
    }

    outputString := fmt.Sprintf(`{
        "tags": {},
        "entries": [
        %s,
        %s
        %s
        ]
    }
    `, tooEarlyJson, rightDayString, tooLateJson)
    return outputString, nil
}

type mockHTTPClient struct{
    errOnDo bool
    bodyOfRequest []byte
}

func (m *mockHTTPClient) Do (req *http.Request) (*http.Response, error) {
    if m.errOnDo {
        return nil, errors.New("new error")
    }

    body, err := io.ReadAll(req.Body)
    if err != nil {
        return nil, err
    }
    m.bodyOfRequest = body
    resp := &http.Response{
        Status: "200 OK",
        StatusCode: 200,
        Body: io.NopCloser(bytes.NewBuffer([]byte{})),
    }
    return resp, nil
}

type mockCommand struct {
    errOnOutput bool
    outputString string
}

func (mc mockCommand) Output() ([]byte, error) {
    if mc.errOnOutput {
        return nil, errors.New("error on output")
    }

    output := []byte(mc.outputString)
    return output, nil
}
