package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/state"
	"github.com/spf13/cobra"
)

func (a *CLI) stateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "State version management",
	}

	cmd.AddCommand(a.stateRollbackCommand())
	cmd.AddCommand(a.stateListCommand())
	cmd.AddCommand(a.stateDeleteCommand())
	cmd.AddCommand(a.stateDownloadCommand())

	return cmd
}

func (a *CLI) stateListCommand() *cobra.Command {
	var opts state.StateVersionListOptions
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List state versions",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			out := cmd.OutOrStdout()

			// first retrieve workspace and current state version so that the
			// user can be informed which state version is current
			workspace, err := a.GetWorkspaceByName(ctx, opts.Organization, opts.Workspace)
			if err != nil {
				return err
			}
			current, err := a.GetCurrentStateVersion(ctx, workspace.ID)
			if errors.Is(err, internal.ErrResourceNotFound) {
				fmt.Fprintln(out, "No state versions found")
				return nil
			}
			if err != nil {
				return err
			}

			for {
				list, err := a.ListStateVersions(cmd.Context())
				if err != nil {
					return err
				}
				for _, sv := range list.Items {
					fmt.Fprintf(out, sv.ID)
					if current.ID == sv.ID {
						fmt.Fprintf(out, " (current)")
					}
					fmt.Fprintln(out)
				}
				if list.NextPage() == nil {
					break
				}
				opts.PageNumber = *list.NextPage()
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Organization, "organization", "", "Name of the organization the workspace belongs to")
	cmd.MarkFlagRequired("organization")

	cmd.Flags().StringVar(&opts.Workspace, "workspace", "", "Name of workspace for which to retreive state versions")
	cmd.MarkFlagRequired("workspace")

	return cmd
}

func (a *CLI) stateDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "delete [id]",
		Short:         "Delete state version",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.DeleteStateVersion(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted state version: %s\n", args[0])
			return nil
		},
	}
}

func (a *CLI) stateDownloadCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "download [id]",
		Short:         "Download state version",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := a.DownloadState(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			var out bytes.Buffer
			if err := json.Indent(&out, state, "", "  "); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), out.String())
			return nil
		},
	}
}

func (a *CLI) stateRollbackCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "rollback [id]",
		Short:         "Rollback to a previous state version",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := a.RollbackStateVersion(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully rolled back state\n")
			return nil
		},
	}
}
