package main

import (
	"fmt"

	"github.com/leg100/go-tfe"
	"github.com/spf13/cobra"
)

func WorkspaceListCommand(config ClientConfig) *cobra.Command {
	var organization string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := config.NewClient()
			if err != nil {
				return err
			}

			list, err := client.Workspaces().List(cmd.Context(), organization, tfe.WorkspaceListOptions{})
			if err != nil {
				return err
			}

			for _, ws := range list.Items {
				fmt.Println(ws.Name)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
