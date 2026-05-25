package cli

import (
	"context"
	"encoding/json"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/workspace"
	workspaceapi "github.com/leg100/otf/internal/workspace/api"
	"github.com/leg100/otf/internal/workspace/execution"

	"github.com/leg100/otf/internal/organization"

	"github.com/leg100/otf/internal/resource"
	"github.com/spf13/cobra"
)

type CLI struct {
	client client
}

type client interface {
	ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
	GetWorkspaceByName(ctx context.Context, organization organization.Name, workspace string) (*workspace.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspaceID resource.TfeID, opts workspace.UpdateOptions) (*workspace.Workspace, error)
	Lock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*workspace.Workspace, error)
	Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*workspace.Workspace, error)
}

func NewCommand(apiClient *otfhttp.Client) *cobra.Command {
	cli := &CLI{}
	cmd := &cobra.Command{
		Use:   "workspaces",
		Short: "Workspace management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.client = &workspaceapi.Client{Client: apiClient}
			return nil
		},
	}

	cmd.AddCommand(cli.workspaceListCommand())
	cmd.AddCommand(cli.workspaceShowCommand())
	cmd.AddCommand(cli.workspaceEditCommand())
	cmd.AddCommand(cli.workspaceLockCommand())
	cmd.AddCommand(cli.workspaceUnlockCommand())

	return cmd
}

func (a *CLI) workspaceListCommand() *cobra.Command {
	var organization organization.Name

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List workspaces",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Workspace], error) {
				return a.client.ListWorkspaces(cmd.Context(), workspace.ListOptions{
					PageOptions:  opts,
					Organization: &organization,
				})
			})
			if err != nil {
				return fmt.Errorf("retrieving existing workspaces: %w", err)
			}
			for _, ws := range list {
				fmt.Fprintln(cmd.OutOrStdout(), ws.Name)
			}

			return nil
		},
	}

	cmd.Flags().Var(&organization, "organization", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}

func (a *CLI) workspaceShowCommand() *cobra.Command {
	var organization organization.Name

	cmd := &cobra.Command{
		Use:           "show [name]",
		Short:         "Show a workspace",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := args[0]

			ws, err := a.client.GetWorkspaceByName(cmd.Context(), organization, workspace)
			if err != nil {
				return err
			}
			out, err := json.MarshalIndent(ws, "", "    ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(out))

			return nil
		},
	}

	cmd.Flags().Var(&organization, "organization", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}

func (a *CLI) workspaceEditCommand() *cobra.Command {
	var (
		organization  organization.Name
		opts          workspace.UpdateOptions
		executionKind string
		poolID        resource.TfeID
		poolIDFlag    = "agent-pool-id"
	)

	cmd := &cobra.Command{
		Use:           "edit [name]",
		Short:         "Edit a workspace",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if executionKind != "" {
				opts.ExecutionKind = new(execution.Kind(executionKind))
			}
			// Set agent pool ID if user set it.
			if cmd.Flags().Changed(poolIDFlag) {
				opts.AgentPoolID = &poolID
			}
			ws, err := a.client.GetWorkspaceByName(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			_, err = a.client.UpdateWorkspace(cmd.Context(), ws.ID, opts)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "updated workspace")
			return nil
		},
	}

	cmd.Flags().StringVarP(&executionKind, "execution-mode", "m", "", "Which execution mode to use. Valid values are remote, local, and agent")
	cmd.Flags().Var(&poolID, poolIDFlag, "ID of the agent pool to use for runs. Required if execution-mode is set to agent.")

	cmd.Flags().Var(&organization, "organization", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}

func (a *CLI) workspaceLockCommand() *cobra.Command {
	var organization organization.Name

	cmd := &cobra.Command{
		Use:           "lock [name]",
		Short:         "Lock a workspace",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := args[0]

			ws, err := a.client.GetWorkspaceByName(cmd.Context(), organization, workspace)
			if err != nil {
				return err
			}
			ws, err = a.client.Lock(cmd.Context(), ws.ID, nil)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully locked workspace %s\n", ws.Name)

			return nil
		},
	}

	cmd.Flags().Var(&organization, "organization", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}

func (a *CLI) workspaceUnlockCommand() *cobra.Command {
	var (
		organization organization.Name
		force        bool
	)

	cmd := &cobra.Command{
		Use:           "unlock [name]",
		Short:         "Unlock a workspace",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := args[0]

			ws, err := a.client.GetWorkspaceByName(cmd.Context(), organization, workspace)
			if err != nil {
				return err
			}
			ws, err = a.client.Unlock(cmd.Context(), ws.ID, nil, force)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully unlocked workspace %s\n", ws.Name)

			return nil
		},
	}

	cmd.Flags().Var(&organization, "organization", "Organization workspace belongs to")
	cmd.Flags().BoolVar(&force, "force", false, "Forceably unlock workspace.")
	cmd.MarkFlagRequired("organization")

	return cmd
}
