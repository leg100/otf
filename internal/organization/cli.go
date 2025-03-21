package organization

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/spf13/cobra"
)

type (
	CLI struct {
		cliService
	}

	// cliService provides the cli with access to organizations
	cliService interface {
		CreateOrganization(ctx context.Context, opts CreateOptions) (*Organization, error)
		Update(ctx context.Context, name resource.OrganizationName, opts UpdateOptions) (*Organization, error)
		Get(ctx context.Context, name resource.OrganizationName) (*Organization, error)
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error)
		DeleteOrganization(ctx context.Context, name resource.OrganizationName) error
		GetEntitlements(ctx context.Context, organization resource.OrganizationName) (Entitlements, error)
		AfterCreateOrganization(hook func(context.Context, *Organization) error)
		BeforeDeleteOrganization(hook func(context.Context, *Organization) error)

		// organization tokens
		CreateToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error)
		// GetOrganizationToken gets the organization token. If a token does not
		// exist, then nil is returned without an error.
		GetOrganizationToken(ctx context.Context, organization resource.OrganizationName) (*OrganizationToken, error)
		DeleteToken(ctx context.Context, organization resource.OrganizationName) error
		WatchOrganizations(context.Context) (<-chan pubsub.Event[*Organization], func())
		getOrganizationTokenByID(ctx context.Context, tokenID resource.TfeID) (*OrganizationToken, error)
	}
)

func NewCommand(client *otfapi.Client) *cobra.Command {
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
			name, err := resource.NewOrganizationName(args[0])
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
