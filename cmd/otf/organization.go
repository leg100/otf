package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/spf13/cobra"
)

func (a *application) organizationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Organization management",
	}

	cmd.AddCommand(a.organizationNewCommand())

	return cmd
}

func (a *application) organizationNewCommand() *cobra.Command {
	opts := organization.OrganizationCreateOptions{}

	cmd := &cobra.Command{
		Use:           "new [name]",
		Short:         "Create a new organization",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = otf.String(args[0])

			org, err := a.CreateOrganization(cmd.Context(), opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created organization %s\n", org.Name)

			return nil
		},
	}

	return cmd
}
