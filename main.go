package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/jm96441n/jrnlNotion/setup"
	"github.com/jm96441n/jrnlNotion/sync"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var Version string

func main() {
    rootFlagSet := flag.NewFlagSet("jrnlNotion", flag.ExitOnError)

    httpClient := &http.Client{}
    cmd := exec.Command("jrnl", "--format", "json")
    entryDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

    syncCommand := sync.NewSyncFlagSet(httpClient, cmd, entryDate)

    setupCommand := setup.NewSetupFlagSet()

    rootCommand := &ffcli.Command{
        ShortUsage: "jrnlNotion [flags] <subcommand>",
        FlagSet: rootFlagSet,
        Subcommands: []*ffcli.Command{setupCommand, syncCommand},
        Exec: func(_ context.Context, args []string) error {
            return flag.ErrHelp
        },
    }

    if err := rootCommand.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
        log.Fatal(err)
    }
}
