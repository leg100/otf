package main

import (
	"encoding/json"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceEditCommand(factory http.ClientFactory) *cobra.Command {
	var (
		spec otf.WorkspaceSpec
		opts otf.WorkspaceUpdateOptions
	)

	cmd := &cobra.Command{
		Use:   "edit [name]",
		Short: "Edit a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec.Name = otf.String(args[0])

			client, err := factory.NewClient()
			if err != nil {
				return err
			}

			ws, err := client.Workspaces().Update(cmd.Context(), spec, opts)
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

	opts.ExecutionMode = cmd.Flags().String("execution-mode", otf.DefaultExecutionMode, "Which execution mode to use. Valid values are remote, local")

	spec.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
