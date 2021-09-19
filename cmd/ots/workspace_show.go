package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func WorkspaceShowCommand(config ClientConfig) *cobra.Command {
	var organization string
	var workspace string

	cmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show a workspace",
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

			out, err := json.MarshalIndent(ws, "", "    ")
			if err != nil {
				return err
			}

			fmt.Println(string(out))

			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
