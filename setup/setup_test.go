package setup_test

import (
	"bytes"
	"context"
	"errors"
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
    reader := &mockStringReader{responsesForReadString: readerResponses}
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
    reader := &mockStringReader{responsesForReadString: readerResponses}
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

func TestSetupReturnsFailedToReadInputErrorIfInputFails(t *testing.T) {
    f := &mockFile{}
    mCC := mockCurrentCronCmd{returnVal: []byte{}}
    mSC := mockSaveCronCmd{}
    cron := setup.NewCron(f, mCC, mSC)
    responses := []string{"databaseid", "notionkey"}
    readerResponses := make([]string, len(responses))
    copy(readerResponses, responses)
    writeBuf := bytes.NewBuffer([]byte{})
    for i := 1; i <= 2; i++ {
        reader := &mockStringReader{responsesForReadString: readerResponses, errOnRead: true, countToErrOn: i}
        c := setup.NewConfig(cron, reader, writeBuf)
        err := c.Exec(context.Background(), []string{})
        if !errors.Is(err, setup.ErrFailedToReadInput) {
            t.Errorf("Expected to get ErrFailedToReadInput as the error for the %d call to read input", i)
        }
    }
}

func TestSetupReturnsErrorWhenSettingCronFails(t *testing.T) {
    testCases := []struct{
        name string
        f *mockFile
        mCC mockCurrentCronCmd
        mSC mockSaveCronCmd
        expectedError error
    }{
        {
            name: "when failing to execute list current crons",
            f: &mockFile{},
            mCC: mockCurrentCronCmd{returnErr: true},
            mSC: mockSaveCronCmd{},
            expectedError: setup.ErrFailedToGetCurrentCron,
        },
        {
            name: "when failing to write new cron file",
            f: &mockFile{errOnWrite: true},
            mCC: mockCurrentCronCmd{},
            mSC: mockSaveCronCmd{},
            expectedError: setup.ErrFailedToWriteNewCronFile,
        },
        {
            name: "when failing to sync writes to cron file",
            f: &mockFile{errOnSync: true},
            mCC: mockCurrentCronCmd{},
            mSC: mockSaveCronCmd{},
            expectedError: setup.ErrFailedToSyncFile,
        },
        {
            name: "when failing to close cron file",
            f: &mockFile{errOnClose: true},
            mCC: mockCurrentCronCmd{},
            mSC: mockSaveCronCmd{},
            expectedError: setup.ErrFailedToCloseFile,
        },
        {
            name: "when failing to execute set new cron command",
            f: &mockFile{},
            mCC: mockCurrentCronCmd{},
            mSC: mockSaveCronCmd{returnErr: true},
            expectedError: setup.ErrFailedToSetNewCron,
        },
    }

    responses := []string{"databaseid", "notionkey"}
    writeBuf := bytes.NewBuffer([]byte{})
    for _, testCase := range testCases {
        readerResponses := make([]string, len(responses))
        copy(readerResponses, responses)
        reader := &mockStringReader{responsesForReadString: readerResponses}

        cron := setup.NewCron(testCase.f, testCase.mCC, testCase.mSC)
        c := setup.NewConfig(cron, reader, writeBuf)
        err := c.Exec(context.Background(), []string{})
        if !errors.Is(err, testCase.expectedError) {
            t.Errorf("Expected error to be %q, got %q", testCase.expectedError, err)
        }
    }

}

type mockStringReader struct{
    responsesForReadString []string
    readCount int
    errOnRead bool
    countToErrOn int
}

func (m *mockStringReader) ReadString(byte) (string, error) {
    resp := m.responsesForReadString[m.readCount]
    m.readCount++
    if m.errOnRead && (m.readCount == m.countToErrOn) {
        return "", errors.New("error")
    }
    return resp, nil
}

type mockFile struct {
    cronCommand string
    errOnWrite bool
    errOnSync bool
    errOnClose bool
}

func (m *mockFile) Sync() error {
    if m.errOnSync {
        return errors.New("error")
    }
    return nil
}

func (m *mockFile) Name() string {
    return "mockName"
}

func (m *mockFile) Write(in []byte) (int, error) {
    if m.errOnWrite {
        return 0, errors.New("error")
    }
    m.cronCommand = string(in)
    return len(in), nil
}

func (m *mockFile) Close() error {
    if m.errOnClose {
        return errors.New("error")
    }
    return nil
}

type mockCurrentCronCmd struct {
    returnErr bool
    returnVal []byte
}

func (m mockCurrentCronCmd) Output() ([]byte, error) {
    if m.returnErr {
        return nil, errors.New("error")
    }
    return m.returnVal, nil
}

type mockSaveCronCmd struct {
    returnErr bool
}

func (m mockSaveCronCmd) Run() error {
    if m.returnErr {
        return errors.New("error")
    }
    return nil
}
