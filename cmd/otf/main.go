package main

import (
	"context"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	if err := Run(ctx, os.Args[1:]); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func Run(ctx context.Context, args []string) error {
	cfg, err := http.NewConfig(LoadCredentials)
	if err != nil {
		return err
	}

	cmd := &cobra.Command{
		Use:           "otf",
		SilenceUsage:  true,
		SilenceErrors: true,
		// Define run func in order to enable cobra's default help functionality
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cmd.PersistentFlags().StringVar(&cfg.Address, "address", http.DefaultAddress, "Address of OTF server")

	cmd.SetArgs(args)

	cmd.AddCommand(OrganizationCommand(cfg))
	cmd.AddCommand(WorkspaceCommand(cfg))
	cmd.AddCommand(AgentCommand(cfg))

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ExecuteContext(ctx); err != nil {
		return err
	}
	return nil
}
