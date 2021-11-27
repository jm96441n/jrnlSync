package setup

import (
	"bytes"
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
        return err
    }
    buf := bytes.NewBuffer(output)
    _, err = buf.Write([]byte(task))
    if err != nil {
        return err
    }

    c.tmpFile.Write(buf.Bytes())
    c.tmpFile.Sync()
    err = c.tmpFile.Close()
    if err != nil {
        return err
    }

    err = c.saveCronCmd.Run()
    if err != nil {
        return fmt.Errorf("failed to save crontab: %w", err)
    }
    return nil
}

