package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/jm96441n/jrnlNotion/cli"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var Version string

func main() {
    rootFlagSet := flag.NewFlagSet("jrnlNotion", flag.ExitOnError)

    httpClient := &http.Client{}
    buf := bytes.NewBuffer([]byte{})
    output := buf
    cmd := exec.Command("jrnl", "--format", "json")
    syncCommand := cli.NewSyncFlagSet(httpClient, output, cmd)

    rootCommand := &ffcli.Command{
        ShortUsage: "jrnlNotion [flags] <subcommand>",
        FlagSet: rootFlagSet,
        Subcommands: []*ffcli.Command{syncCommand},
        Exec: func(_ context.Context, args []string) error {
            return flag.ErrHelp
        },
    }

    if err := rootCommand.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
        log.Fatal(err)
    }
}
