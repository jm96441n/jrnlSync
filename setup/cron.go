package setup

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Cron struct {
    tmpFile fileWriterSyncer
    currentCronCmd commandOutputter
    saveCronCmd commandRunner
}

type fileWriterSyncer interface {
    Sync() error
    Name() string
    io.WriteCloser
}

type commandOutputter interface {
    Output() ([]byte, error)
}

type commandRunner interface {
    Run() error
}

var ErrFailedToGetCurrentCron = errors.New("failed to get current cron list")
var ErrFailedToWriteNewCronTaskToBuffer = errors.New("failed to write the new cron task to the buffer")
var ErrFailedToWriteNewCronFile = errors.New("failed to write the new cron tmp file")
var ErrFailedToSyncFile = errors.New("failed to sync writes to the new cron file")
var ErrFailedToCloseFile = errors.New("failed to close the cron tmp file")
var ErrFailedToSetNewCron = errors.New("failed to set new cron")

func NewCron(tmpFile fileWriterSyncer, currentCronCmd commandOutputter, saveCronCmd commandRunner) Cron {
    return Cron{
        tmpFile: tmpFile,
        saveCronCmd: saveCronCmd,
        currentCronCmd: currentCronCmd,
    }
}

func (c Cron) addCron(task string) error {
    output, err := c.currentCronCmd.Output()
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToGetCurrentCron, err)
    }
    buf := bytes.NewBuffer(output)
    _, err = buf.Write([]byte(task))
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToWriteNewCronTaskToBuffer, err)
    }

    _, err = c.tmpFile.Write(buf.Bytes())
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToWriteNewCronFile, err)
    }

    err = c.tmpFile.Sync()
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToSyncFile, err)
    }
    err = c.tmpFile.Close()
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToCloseFile, err)
    }

    err = c.saveCronCmd.Run()
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToSetNewCron, err)
    }
    return nil
}

