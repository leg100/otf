package main

import (
	"fmt"

	"github.com/leg100/otf/tokens"
	"github.com/spf13/cobra"
)

func (a *application) agentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Agent management",
	}

	cmd.AddCommand(a.agentTokenCommand())

	return cmd
}

func (a *application) agentTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Agent token management",
	}

	cmd.AddCommand(a.agentTokenNewCommand())

	return cmd
}

func (a *application) agentTokenNewCommand() *cobra.Command {
	opts := tokens.CreateAgentTokenOptions{}

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
