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

	if err := Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Run(ctx context.Context, args []string) error {
	cmd := &cobra.Command{
		Use:           "ots",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetArgs(args)

	cmd.AddCommand(LoginCommand(&SystemDirectories{}))
	cmd.AddCommand(OrganizationCommand())
	cmd.AddCommand(WorkspaceCommand())

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ExecuteContext(ctx); err != nil {
		return err
	}
	return nil
}
