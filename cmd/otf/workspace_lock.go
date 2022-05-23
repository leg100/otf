package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceLockCommand(factory http.ClientFactory) *cobra.Command {
	var spec otf.WorkspaceSpec

	cmd := &cobra.Command{
		Use:   "lock [name]",
		Short: "Lock a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec.Name = otf.String(args[0])

			client, err := factory.NewClient()
			if err != nil {
				return err
			}

			// API only provides the ability to lock workspace by id so we need
			// to fetch that first.
			ws, err := client.Workspaces().Get(cmd.Context(), spec)
			if err != nil {
				return err
			}

			_, err = client.Workspaces().Lock(cmd.Context(), otf.WorkspaceSpec{ID: &ws.ID}, otf.WorkspaceLockOptions{})
			if err != nil {
				return err
			}

			fmt.Printf("Successfully locked workspace %s\n", ws.Name())

			return nil
		},
	}

	spec.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
