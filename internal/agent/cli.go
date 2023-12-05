package agent

import (
	"context"
	"fmt"

	otfapi "github.com/leg100/otf/internal/api"

	"github.com/spf13/cobra"
)

type (
	agentCLI struct {
		agentCLIService
	}

	agentCLIService interface {
		CreateAgentToken(ctx context.Context, poolID string, opts CreateAgentTokenOptions) (*agentToken, []byte, error)
	}
)

func NewAgentsCommand(client *otfapi.Client) *cobra.Command {
	cli := &agentCLI{}
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Agent management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.agentCLIService = &rpcClient{Client: client}
			return nil
		},
	}

	cmd.AddCommand(cli.agentTokenCommand())

	return cmd
}

func (a *agentCLI) agentTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Agent token management",
	}

	cmd.AddCommand(a.agentTokenNewCommand())

	return cmd
}

func (a *agentCLI) agentTokenNewCommand() *cobra.Command {
	var (
		poolID string
		opts   = CreateAgentTokenOptions{}
	)
	cmd := &cobra.Command{
		Use:           "new",
		Short:         "Create a new agent token",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, token, err := a.CreateAgentToken(cmd.Context(), poolID, opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created agent token: %s\n", token)

			return nil
		},
	}
	cmd.Flags().StringVar(&poolID, "agent-pool-id", "", "ID of the agent pool in which the token is to be created.")
	cmd.MarkFlagRequired("agent-pool-id")

	cmd.Flags().StringVar(&opts.Description, "description", "", "Provide a description for the token.")
	cmd.MarkFlagRequired("description")

	return cmd
}
