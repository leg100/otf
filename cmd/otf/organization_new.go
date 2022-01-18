package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func OrganizationNewCommand(factory http.ClientFactory) *cobra.Command {
	opts := otf.OrganizationCreateOptions{}

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = otf.String(args[0])

			client, err := factory.NewClient()
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

	return cmd
}
