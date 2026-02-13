package organization

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"

	"github.com/spf13/cobra"
)

type (
	CLI struct {
		cliService
	}

	// cliService provides the cli with access to organizations
	cliService interface {
		CreateOrganization(ctx context.Context, opts CreateOptions) (*Organization, error)
		DeleteOrganization(ctx context.Context, name Name) error
	}
)

func NewCommand(client *otfhttp.Client) *cobra.Command {
	cli := &CLI{}
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Organization management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.cliService = &Client{Client: client}
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
				Name: new(args[0]),
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
			name, err := NewName(args[0])
			if err != nil {
				return err
			}
			if err := a.DeleteOrganization(cmd.Context(), name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted organization %s\n", args[0])
			return nil
		},
	}
}
