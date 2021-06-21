package main

import (
	"context"
	"fmt"
	"os"

	cmdutil "github.com/leg100/ots/cmd"
	"github.com/spf13/cobra"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	cmd := &cobra.Command{
		Use:           "ots",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(LoginCommand(&SystemDirectories{}))
	cmd.AddCommand(OrganizationCommand())

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
