package setup

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
    Reader stringReader
    Cron Cron
}

type stringReader interface {
    ReadString(byte) (string, error)
}

func NewSetupFlagSet(cron Cron) *ffcli.Command {
    c := Config{Reader: bufio.NewReader(os.Stdin), Cron: cron}
    setupFlagSet := flag.NewFlagSet("jrnlSync setup", flag.ExitOnError)

    return &ffcli.Command{
        Name:       "setup",
        ShortUsage: "jrnlSync setup",
        ShortHelp:  "Sets up the daily cron job to sync your notes to notion",
        FlagSet:    setupFlagSet,
        Exec:       c.Exec,
    }
}

func (c Config) Exec(_ context.Context, _ []string) error {
    fmt.Printf("Please enter the DB id that the notes will be synced to: ")
    dbid, err := c.Reader.ReadString('\n')
    if err != nil {
        return err
    }
    dbid = strings.ReplaceAll(dbid, "\n", "")
    fmt.Printf("Please enter your notion integration key: ")
    notionKey, err := c.Reader.ReadString('\n')
    if err != nil {
        return err
    }
    notionKey = strings.ReplaceAll(notionKey, "\n", "")

    fmt.Println("Scheduling cron task to run every night to sync")

    err = c.Cron.addCron(fmt.Sprintf("1 12 * * * jrnlSync notion -d %s -k %s > ~/.jrnlSyncLogs.txt 2>&1\n", dbid, notionKey))
    if err != nil {
        return err
    }

    fmt.Println("Scheduled to run every day right after midinight!")

    return nil
}
