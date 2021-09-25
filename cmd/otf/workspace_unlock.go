package main

import (
	"fmt"

	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceUnlockCommand(factory http.ClientFactory) *cobra.Command {
	var organization string
	var workspace string

	cmd := &cobra.Command{
		Use:   "unlock [name]",
		Short: "Unlock a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace = args[0]

			client, err := factory.NewClient()
			if err != nil {
				return err
			}

			ws, err := client.Workspaces().Read(cmd.Context(), organization, workspace)
			if err != nil {
				return err
			}

			ws, err = client.Workspaces().Unlock(cmd.Context(), ws.ID)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully locked workspace %s\n", ws.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
