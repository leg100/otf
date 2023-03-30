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

	cmd.AddCommand(a.newOrganizationCommand())
	cmd.AddCommand(a.deleteOrganizationCommand())

	return cmd
}

func (a *application) newOrganizationCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "new [name]",
		Short:         "Create a new organization",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			org, err := a.CreateOrganization(cmd.Context(), organization.OrganizationCreateOptions{
				Name: otf.String(args[0]),
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created organization %s\n", org.Name)

			return nil
		},
	}
}

func (a *application) deleteOrganizationCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "delete [organization]",
		Short:         "Delete an organization",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.DeleteOrganization(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted organization %s\n", args[0])
			return nil
		},
	}
}
