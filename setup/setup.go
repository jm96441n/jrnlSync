package setup

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
    reader stringReader
}

type stringReader interface {
    ReadString(byte) (string, error)
}

func NewSetupFlagSet() *ffcli.Command {
    c := Config{reader: bufio.NewReader(os.Stdin)}
    setupFlagSet := flag.NewFlagSet("jrnlNotion setup", flag.ExitOnError)

    return &ffcli.Command{
        Name:       "setup",
        ShortUsage: "jrnlNotion setup",
        ShortHelp:  "Sets up the daily cron job to sync your notes to notion",
        FlagSet:    setupFlagSet,
        Exec:       c.Exec,
    }
}

func (c Config) Exec(_ context.Context, _ []string) error {
    fmt.Printf("Please enter the DB id that the notes will be synced to: ")
    dbid, err := c.reader.ReadString('\n')
    if err != nil {
        return err
    }
    dbid = strings.ReplaceAll(dbid, "\n", "")
    fmt.Printf("Please enter your notion integration key: ")
    notionKey, err := c.reader.ReadString('\n')
    if err != nil {
        return err
    }
    notionKey = strings.ReplaceAll(notionKey, "\n", "")

    fmt.Println("Scheduling cron task to run every night to sync")
    f, err := os.CreateTemp("", "notion_cron")
    defer os.Remove(f.Name())
    if err != nil {
        return err
    }
    getCurrentCronttabCmd := exec.Command(
        "crontab",
        "-l",
    )
    output, err := getCurrentCronttabCmd.Output()
    if err != nil {
        return err
    }
    buf := bytes.NewBuffer(output)
    notionCron := fmt.Sprintf("1 12 * * * jrnlNotion -d %s -k %s > ~/.jrnlToNotionLogs.txt 2>&1\n", dbid, notionKey)
    _, err = buf.Write([]byte(notionCron))
    if err != nil {
        return err
    }

    f.Write(buf.Bytes())
    f.Sync()
    filename := f.Name()
    err = f.Close()
    if err != nil {
        return err
    }

    saveCrontabCmd := exec.Command("crontab", filename)
    err = saveCrontabCmd.Run()
    if err != nil {
        return fmt.Errorf("failed to save crontab: %w", err)
    }


    fmt.Println("Scheduled to run every day right after midinight!")

    return nil
}
