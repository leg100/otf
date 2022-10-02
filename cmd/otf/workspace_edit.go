package main

import (
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func WorkspaceEditCommand(factory http.ClientFactory) *cobra.Command {
	var (
		spec otf.WorkspaceSpec
		opts otf.WorkspaceUpdateOptions

		mode *string
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

			if mode != nil && *mode != "" {
				opts.ExecutionMode = (*otf.ExecutionMode)(mode)
			}

			ws, err := client.UpdateWorkspace(cmd.Context(), spec, opts)
			if err != nil {
				return err
			}

			if opts.ExecutionMode != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "updated execution mode: %s\n", ws.ExecutionMode())
			}

			return nil
		},
	}

	mode = cmd.Flags().StringP("execution-mode", "m", "", "Which execution mode to use. Valid values are remote, local, and agent")

	spec.OrganizationName = cmd.Flags().String("organization", "", "Organization workspace belongs to")
	cmd.MarkFlagRequired("organization")

	return cmd
}
