package cli

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/runner"
	runnerapi "github.com/leg100/otf/internal/runner/api"

	"github.com/leg100/otf/internal/resource"

	"github.com/spf13/cobra"
)

type (
	cli struct {
		client
	}

	client interface {
		CreateAgentToken(ctx context.Context, poolID resource.TfeID, opts runner.CreateAgentTokenOptions) (*runner.AgentToken, []byte, error)
		Register(ctx context.Context, opts runner.RegisterRunnerOptions) (*runner.RunnerMeta, error)
		AwaitAllocatedJobs(ctx context.Context, agentID resource.TfeID) ([]*runner.Job, error)
		UpdateStatus(ctx context.Context, agentID resource.TfeID, status runner.RunnerStatus) error
		StartJob(ctx context.Context, jobID resource.TfeID) ([]byte, error)
	}
)

func NewAgentsCommand(apiClient *otfhttp.Client) *cobra.Command {
	cli := &cli{}
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Agent management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.client = &runnerapi.Client{Client: apiClient}
			return nil
		},
	}

	cmd.AddCommand(cli.agentTokenCommand())

	return cmd
}

func (a *cli) agentTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Agent token management",
	}

	cmd.AddCommand(a.agentTokenNewCommand())

	return cmd
}

func (a *cli) agentTokenNewCommand() *cobra.Command {
	var (
		poolID resource.TfeID
		opts   = runner.CreateAgentTokenOptions{}
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
	cmd.Flags().Var(&poolID, "agent-pool-id", "ID of the agent pool in which the token is to be created.")
	cmd.MarkFlagRequired("agent-pool-id")

	cmd.Flags().StringVar(&opts.Description, "description", "", "Provide a description for the token.")
	cmd.MarkFlagRequired("description")

	return cmd
}
