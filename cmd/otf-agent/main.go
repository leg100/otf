package main

import (
	"context"
	"fmt"
	"os"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
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
	var help bool

	cfg, err := http.NewConfig()
	if err != nil {
		return err
	}

	cmd := &cobra.Command{
		Use:           "otf-agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		// Define run func in order to enable cobra's default help functionality
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().BoolVarP(&help, "help", "h", false, "Print usage information")
	cmd.Flags().StringVar(&cfg.Address, "address", http.DefaultAddress, "Address of OTF server")
	cmd.Flags().StringVar(&cfg.Token, "token", "", "Agent token for authentication")
	cmd.MarkFlagRequired("token")

	loggerCfg := cmdutil.NewLoggerConfigFromFlags(cmd.Flags())

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		return err
	}

	if help {
		if err := cmd.Help(); err != nil {
			return err
		}
		return nil
	}

	logger, err := cmdutil.NewLogger(loggerCfg)
	if err != nil {
		return err
	}

	// NewClient sends unauthenticated ping to server
	client, err := cfg.NewClient()
	if err != nil {
		return err
	}

	// Confirm token validity
	at, err := client.GetAgentToken(ctx, "")
	if err != nil {
		return fmt.Errorf("attempted authentication: %w", err)
	}
	logger.Info("successfully authenticated", "organization", at.OrganizationName(), "token_id", at.ID())

	agent, err := agent.NewAgent(logger, client, agent.NewAgentOptions{
		Organization: otf.String(at.OrganizationName()),
		Mode:         agent.ExternalAgentMode,
	})
	if err != nil {
		return fmt.Errorf("unable to start agent: %w", err)
	}
	// blocks
	return agent.Start(ctx)
}
