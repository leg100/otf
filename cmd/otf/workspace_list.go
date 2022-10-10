package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceListCommand(factory http.ClientFactory) *cobra.Command {
	var opts otf.WorkspaceListOptions

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List workspaces",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := factory.NewClient()
			if err != nil {
				return err
			}

			for {
				list, err := client.ListWorkspaces(cmd.Context(), opts)
				if err != nil {
					return err
				}
				for _, ws := range list.Items {
					fmt.Fprintln(cmd.OutOrStdout(), ws.Name())
				}
				if list.NextPage() != nil {
					opts.PageNumber = *list.NextPage()
				} else {
					break
				}
			}

			return nil
		},
	}

	opts.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
