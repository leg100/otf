package main

import (
	"context"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal/cli"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	if err := (&cli.CLI{}).Run(ctx, os.Args[1:], os.Stdout); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}
