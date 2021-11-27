package setup

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
    Notion bool
    reader stringReader
    cron Cron
    out io.Writer
}

type stringReader interface {
    ReadString(byte) (string, error)
}

var ErrFailedToReadInput = errors.New("failed to read input from user")

func NewSetupFlagSet(cron Cron) *ffcli.Command {
    c := NewConfig(cron, bufio.NewReader(os.Stdin), os.Stdout)
    setupFlagSet := flag.NewFlagSet("jrnlSync setup", flag.ExitOnError)
    setupFlagSet.BoolVar(&c.Notion, "n", false, "setup syncing to notion")

    return &ffcli.Command{
        Name:       "setup",
        ShortUsage: "jrnlSync setup [flags]",
        ShortHelp:  "Sets up the daily cron job to sync your notes to notion",
        FlagSet:    setupFlagSet,
        Exec:       c.Exec,
    }
}

func NewConfig(cron Cron, reader stringReader, out io.Writer) Config {
    return  Config{
        reader: reader,
        cron: cron,
        out: out,
    }
}

func (c Config) Exec(_ context.Context, _ []string) error {
    fmt.Fprint(c.out, "Please enter the DB id that the notes will be synced to: ")
    dbid, err := c.reader.ReadString('\n')
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToReadInput, err)
    }
    dbid = strings.ReplaceAll(dbid, "\n", "")
    fmt.Printf("Please enter your notion integration key: ")
    notionKey, err := c.reader.ReadString('\n')
    if err != nil {
        return fmt.Errorf("%w: %s", ErrFailedToReadInput, err)
    }
    notionKey = strings.ReplaceAll(notionKey, "\n", "")

    fmt.Fprint(c.out, "Scheduling cron task to run every night to sync")

    err = c.cron.addCron(fmt.Sprintf("1 12 * * * jrnlSync notion -d %s -k %s > ~/.jrnlSyncLogs.txt 2>&1\n", dbid, notionKey))
    if err != nil {
        return err
    }

    fmt.Fprint(c.out, "Scheduled to run every day right after midinight!")

    return nil
}
