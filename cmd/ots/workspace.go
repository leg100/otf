package main

import (
	"context"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

func WorkspaceCommand() *cobra.Command {
	cfg := clientConfig{}

	cmd := &cobra.Command{
		Use:   "workspaces",
		Short: "Workspace management",
	}
	cmd.Flags().StringVar(&cfg.Address, "address", DefaultAddress, "Address of OTS server")
	cmd.Flags().StringVar(&cfg.Token, "token", os.Getenv("OTS_TOKEN"), "Authentication token")

	cmd.AddCommand(WorkspaceLockCommand(&cfg))
	//cmd.AddCommand(WorkspaceUnlockCommand(&cfg))

	return cmd
}

type FakeWorkspacesClient struct {
	tfe.Workspaces
}

func (f *FakeWorkspacesClient) Read(ctx context.Context, org string, ws string) (*tfe.Workspace, error) {
	return &tfe.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) Lock(ctx context.Context, id string, opts tfe.WorkspaceLockOptions) (*tfe.Workspace, error) {
	return &tfe.Workspace{
		ID: "ws-123",
	}, nil
}
