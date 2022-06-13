package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceListCommand(factory http.ClientFactory) *cobra.Command {
	var (
		opts otf.WorkspaceListOptions
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := factory.NewClient()
			if err != nil {
				return err
			}

			list, err := client.Workspaces().List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			for _, ws := range list.Items {
				fmt.Println(ws.Name())
			}

			return nil
		},
	}

	opts.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
