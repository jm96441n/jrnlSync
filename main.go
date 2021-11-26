package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/jm96441n/jrnlSync/setup"
	"github.com/jm96441n/jrnlSync/sync"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var Version string

func main() {
    rootFlagSet := flag.NewFlagSet("jrnlSync", flag.ExitOnError)

    httpClient := &http.Client{}
    jrnlCmd := exec.Command("jrnl", "--format", "json")
    entryDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
    notionSyncCommand := sync.NewNotionSyncFlagSet(httpClient, jrnlCmd, entryDate)

    cronTmpFile, err := os.CreateTemp("", "jrnlSync")
    if err != nil {
        log.Fatal(err)
    }
    defer os.Remove(cronTmpFile.Name())
    getCurrentCronttabCmd := exec.Command(
        "crontab",
        "-l",
    )

    cron := setup.NewCron(cronTmpFile ,getCurrentCronttabCmd)


    setupCommand := setup.NewSetupFlagSet(cron)

    rootCommand := &ffcli.Command{
        ShortUsage: "jrnlSync [flags] <subcommand>",
        FlagSet: rootFlagSet,
        Subcommands: []*ffcli.Command{setupCommand, notionSyncCommand},
        Exec: func(_ context.Context, args []string) error {
            return flag.ErrHelp
        },
    }

    if err := rootCommand.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
        log.Fatal(err)
    }
}
