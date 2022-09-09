package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func AgentTokenNewCommand(factory http.ClientFactory) *cobra.Command {
	opts := otf.AgentTokenCreateOptions{}

	cmd := &cobra.Command{
		Use:   "new [description]",
		Short: "Create a new agent token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Description = args[0]

			client, err := factory.NewClient()
			if err != nil {
				return err
			}

			at, err := client.CreateAgentToken(cmd.Context(), opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created agent token: %s\n", *at.Token())

			return nil
		},
	}
	cmd.Flags().StringVar(&opts.OrganizationName, "organization", "", "Organization in which to create agent token.")
	cmd.MarkFlagRequired("organization")

	return cmd
}
