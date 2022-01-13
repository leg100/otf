package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceLockCommand(factory http.ClientFactory) *cobra.Command {
	var spec otf.WorkspaceSpecifier

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

			ws, err := client.Workspaces().Get(cmd.Context(), spec)
			if err != nil {
				return err
			}

			ws, err = client.Workspaces().Lock(cmd.Context(), spec, otf.WorkspaceLockOptions{})
			if err != nil {
				return err
			}

			fmt.Printf("Successfully locked workspace %s\n", ws.Name)

			return nil
		},
	}

	spec.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
