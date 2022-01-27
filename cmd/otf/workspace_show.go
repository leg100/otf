package main

import (
	"encoding/json"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceShowCommand(factory http.ClientFactory) *cobra.Command {
	var spec otf.WorkspaceSpec

	cmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show a workspace",
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

			out, err := json.MarshalIndent(ws, "", "    ")
			if err != nil {
				return err
			}

			fmt.Println(string(out))

			return nil
		},
	}

	spec.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
