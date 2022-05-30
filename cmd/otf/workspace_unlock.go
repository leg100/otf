package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceUnlockCommand(factory http.ClientFactory) *cobra.Command {
	var spec otf.WorkspaceSpec

	cmd := &cobra.Command{
		Use:   "unlock [name]",
		Short: "Unlock a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec.Name = otf.String(args[0])

			client, err := factory.NewClient()
			if err != nil {
				return err
			}
			ws, err := client.Workspaces().Unlock(cmd.Context(), spec, otf.WorkspaceUnlockOptions{
				Requestor: &otf.AnonymousUser,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Successfully unlocked workspace %s\n", ws.Name())

			return nil
		},
	}

	spec.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
