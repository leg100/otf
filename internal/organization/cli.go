package organization

import (
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http"
	"github.com/spf13/cobra"
)

type CLI struct {
	Service
}

func NewCommand(httpClient *http.Client) *cobra.Command {
	cli := &CLI{}
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Organization management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.Service = &Client{JSONAPIClient: httpClient}
			return nil
		},
	}
	cmd.AddCommand(cli.newOrganizationCommand())
	cmd.AddCommand(cli.deleteOrganizationCommand())

	return cmd
}

func (a *CLI) newOrganizationCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "new [name]",
		Short:         "Create a new organization",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			org, err := a.CreateOrganization(cmd.Context(), CreateOptions{
				Name: internal.String(args[0]),
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created organization %s\n", org.Name)

			return nil
		},
	}
}

func (a *CLI) deleteOrganizationCommand() *cobra.Command {
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
