package main

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
	"github.com/spf13/cobra"
)

func OrganizationNewCommand(config ClientConfig) *cobra.Command {
	opts := tfe.OrganizationCreateOptions{}

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = ots.String(args[0])

			client, err := config.NewClient()
			if err != nil {
				return err
			}

			org, err := client.Organizations().Create(cmd.Context(), opts)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully created organization %s\n", org.Name)

			return nil
		},
	}

	opts.Email = cmd.Flags().String("email", "", "Email of the owner of the organization")
	cmd.MarkFlagRequired("email")

	return cmd
}
