package main

import (
	"fmt"

	"github.com/leg100/otf/auth"
	"github.com/spf13/cobra"
)

func (a *application) teamCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teams",
		Short: "Team management",
	}

	cmd.AddCommand(a.teamNewCommand())
	cmd.AddCommand(a.teamDeleteCommand())

	return cmd
}

func (a *application) teamNewCommand() *cobra.Command {
	var organization string

	cmd := &cobra.Command{
		Use:           "new [name]",
		Short:         "Create a new team",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.CreateTeam(cmd.Context(), auth.NewTeamOptions{
				Organization: organization,
				Name:         args[0],
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created team %s\n", team.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")

	return cmd
}

func (a *application) teamDeleteCommand() *cobra.Command {
	var organization string

	cmd := &cobra.Command{
		Use:           "delete [name]",
		Short:         "Delete a team",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.GetTeam(cmd.Context(), organization, args[0])
			if err != nil {
				return err
			}
			if err := a.DeleteTeam(cmd.Context(), team.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted team %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")

	return cmd
}
