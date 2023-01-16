package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceUnlockCommand(factory http.ClientFactory) *cobra.Command {
	var organization string

	cmd := &cobra.Command{
		Use:           "unlock [name]",
		Short:         "Unlock a workspace",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := args[0]

			client, err := factory.NewClient()
			if err != nil {
				return err
			}
			ws, err := client.GetWorkspaceByName(cmd.Context(), organization, workspace)
			if err != nil {
				return err
			}
			ws, err = client.UnlockWorkspace(cmd.Context(), ws.ID(), otf.WorkspaceUnlockOptions{})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully unlocked workspace %s\n", ws.Name())

			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
