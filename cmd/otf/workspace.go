package main

import (
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceCommand(factory http.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspaces",
		Short: "Workspace management",
	}

	cmd.AddCommand(WorkspaceListCommand(factory))
	cmd.AddCommand(WorkspaceShowCommand(factory))
	cmd.AddCommand(WorkspaceEditCommand(factory))
	cmd.AddCommand(WorkspaceLockCommand(factory))
	cmd.AddCommand(WorkspaceUnlockCommand(factory))

	return cmd
}
