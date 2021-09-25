package main

import (
	"os"

	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceCommand() *cobra.Command {
	cfg := http.ClientConfig{}

	cmd := &cobra.Command{
		Use:   "workspaces",
		Short: "Workspace management",
	}
	cmd.Flags().StringVar(&cfg.Address, "address", http.DefaultAddress, "Address of OTF server")
	cmd.Flags().StringVar(&cfg.Token, "token", os.Getenv("OTF_TOKEN"), "Authentication token")

	cmd.AddCommand(WorkspaceListCommand(&cfg))
	cmd.AddCommand(WorkspaceShowCommand(&cfg))
	cmd.AddCommand(WorkspaceEditCommand(&cfg))
	cmd.AddCommand(WorkspaceLockCommand(&cfg))
	cmd.AddCommand(WorkspaceUnlockCommand(&cfg))

	return cmd
}
