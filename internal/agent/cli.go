package agent

import (
	"fmt"

	otfapi "github.com/leg100/otf/internal/api"

	"github.com/spf13/cobra"
)

type CLI struct {
	Service
}

func NewAgentsCommand(api *otfapi.Client) *cobra.Command {
	cli := &CLI{}
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Agent management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.Service = &rpcClient{Client: api}
			return nil
		},
	}

	cmd.AddCommand(cli.agentTokenCommand())

	return cmd
}

func (a *CLI) agentTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Agent token management",
	}

	cmd.AddCommand(a.agentTokenNewCommand())

	return cmd
}

func (a *CLI) agentTokenNewCommand() *cobra.Command {
	opts := CreateAgentTokenOptions{}

	cmd := &cobra.Command{
		Use:           "new [description]",
		Short:         "Create a new agent token",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Description = args[0]

			token, err := a.CreateAgentToken(cmd.Context(), opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created agent token: %s\n", token)

			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Organization, "organization", "", "Organization in which to create agent token.")
	cmd.MarkFlagRequired("organization")

	return cmd
}
