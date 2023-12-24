package agent

import (
	"context"
	"encoding/json"
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
		CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error)
	}
)

func NewAgentsCommand(apiClient *otfapi.Client) *cobra.Command {
	cli := &agentCLI{}
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Agent management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Root().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.agentCLIService = &client{Client: apiClient}
			return nil
		},
	}

	cmd.AddCommand(cli.agentTokenCommand())
	cmd.AddCommand(cli.agentPoolCommand())

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

			fmt.Fprint(cmd.OutOrStdout(), string(token))

			return nil
		},
	}
	cmd.Flags().StringVar(&poolID, "agent-pool-id", "", "ID of the agent pool in which the token is to be created.")
	cmd.MarkFlagRequired("agent-pool-id")

	cmd.Flags().StringVar(&opts.Description, "description", "", "Provide a description for the token.")
	cmd.MarkFlagRequired("description")

	return cmd
}

func (a *agentCLI) agentPoolCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Agent pool management",
	}
	cmd.AddCommand(a.agentPoolNewCommand())
	return cmd
}

func (a *agentCLI) agentPoolNewCommand() *cobra.Command {
	var opts CreateAgentPoolOptions
	cmd := &cobra.Command{
		Use:           "new",
		Short:         "Create a new agent pool",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pool, err := a.CreateAgentPool(cmd.Context(), opts)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(pool, "", "\t")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of agent pool. Required")
	cmd.MarkFlagRequired("name")

	cmd.Flags().StringVar(&opts.Organization, "organization", "", "Agent pool's organization. Required")
	cmd.MarkFlagRequired("organization")

	return cmd
}
