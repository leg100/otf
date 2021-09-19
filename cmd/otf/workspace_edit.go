package main

import (
	"encoding/json"
	"fmt"

	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
	"github.com/spf13/cobra"
)

func WorkspaceEditCommand(config ClientConfig) *cobra.Command {
	var organization, workspace string

	var opts tfe.WorkspaceUpdateOptions

	cmd := &cobra.Command{
		Use:   "edit [name]",
		Short: "Edit a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace = args[0]

			client, err := config.NewClient()
			if err != nil {
				return err
			}

			ws, err := client.Workspaces().Update(cmd.Context(), organization, workspace, opts)
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

	cmd.Flags().StringVar(&organization, "organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
