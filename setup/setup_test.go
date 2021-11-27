package setup_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/jm96441n/jrnlSync/setup"
)

func TestCronIsSetWhenThereAreNoCrons(t *testing.T) {
    f := &mockFile{}
    mCC := mockCurrentCronCmd{}
    mSC := mockSaveCronCmd{}
    cron := setup.NewCron(f, mCC, mSC)
    responses := []string{"databaseid", "notionkey"}
    readerResponses := make([]string, len(responses))
    copy(readerResponses, responses)
    reader := &mockStringReader{readerResponses}
    writeBuf := bytes.NewBuffer([]byte{})
    c := setup.NewConfig(cron, reader, writeBuf)
    err := c.Exec(context.Background(), []string{})
    if err != nil {
        t.Error(err)
    }
    expectedCron := fmt.Sprintf("1 12 * * * jrnlSync notion -d %s -k %s > ~/.jrnlSyncLogs.txt 2>&1\n", responses[0], responses[1])
    if expectedCron != f.cronCommand {
        t.Errorf("Expected cron string to be %q, got %q", expectedCron, f.cronCommand)
    }
}

func TestCronIsSetWhenAppendingToExistingCrons(t *testing.T) {
    f := &mockFile{}
    existingCron := "* * * * * echo \"hello world\"\n"
    mCC := mockCurrentCronCmd{returnVal: []byte(existingCron)}
    mSC := mockSaveCronCmd{}
    cron := setup.NewCron(f, mCC, mSC)
    responses := []string{"databaseid", "notionkey"}
    readerResponses := make([]string, len(responses))
    copy(readerResponses, responses)
    reader := &mockStringReader{readerResponses}
    writeBuf := bytes.NewBuffer([]byte{})
    c := setup.NewConfig(cron, reader, writeBuf)
    err := c.Exec(context.Background(), []string{})
    if err != nil {
        t.Error(err)
    }
    expectedCron := fmt.Sprintf("%s1 12 * * * jrnlSync notion -d %s -k %s > ~/.jrnlSyncLogs.txt 2>&1\n", existingCron, responses[0], responses[1])
    if expectedCron != f.cronCommand {
        t.Errorf("Expected cron string to be %q, got %q", expectedCron, f.cronCommand)
    }
}

type mockStringReader struct{
    responsesForReadString []string
}

func (m *mockStringReader) ReadString(byte) (string, error) {
    resp := m.responsesForReadString[0]
    m.responsesForReadString = m.responsesForReadString[1:]
    return resp, nil
}

type mockFile struct {
    cronCommand string
}

func (m *mockFile) Sync() error {
    return nil
}

func (m *mockFile) Name() string {
    return "mockName"
}

func (m *mockFile) Write(in []byte) (int, error) {
    m.cronCommand = string(in)
    return len(in), nil
}

func (m *mockFile) Close() error {
    return nil
}

type mockCurrentCronCmd struct {
    returnVal []byte
}

func (m mockCurrentCronCmd) Output() ([]byte, error) {
    return m.returnVal, nil
}

type mockSaveCronCmd struct {}

func (m mockSaveCronCmd) Run() error {
    return nil
}
