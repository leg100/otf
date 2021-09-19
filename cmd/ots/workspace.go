package main

import (
	"os"

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

	cmd.AddCommand(WorkspaceListCommand(&cfg))
	cmd.AddCommand(WorkspaceShowCommand(&cfg))
	cmd.AddCommand(WorkspaceEditCommand(&cfg))
	cmd.AddCommand(WorkspaceLockCommand(&cfg))
	cmd.AddCommand(WorkspaceUnlockCommand(&cfg))

	return cmd
}
