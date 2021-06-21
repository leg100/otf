package main

import (
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

func WorkspaceLockCommand(config ClientConfig) *cobra.Command {
	var organization string
	var workspace string

	cmd := &cobra.Command{
		Use:   "lock [name]",
		Short: "Lock a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace = args[0]

			client, err := config.NewClient()
			if err != nil {
				return err
			}

			ws, err := client.Workspaces().Read(cmd.Context(), organization, workspace)
			if err != nil {
				return err
			}

			ws, err = client.Workspaces().Lock(cmd.Context(), ws.ID, tfe.WorkspaceLockOptions{})
			if err != nil {
				return err
			}

			fmt.Printf("Successfully locked workspace %s\n", ws.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "Organization workspace belongs to")

	return cmd
}
