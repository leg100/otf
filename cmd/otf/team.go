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
	cmd.AddCommand(a.addTeamMembershipCommand())
	cmd.AddCommand(a.deleteTeamMembershipCommand())

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

func (a *application) addTeamMembershipCommand() *cobra.Command {
	var (
		organization string
		name         string // team name
	)

	cmd := &cobra.Command{
		Use:           "add-user [username]",
		Short:         "Add user to team",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.GetTeam(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			if err := a.AddTeamMembership(cmd.Context(), args[0], team.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully added %s to %s\n", args[0], name)
			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")
	cmd.Flags().StringVar(&name, "team", "", "Team in which to perform action")
	cmd.MarkFlagRequired("team")

	return cmd
}

func (a *application) deleteTeamMembershipCommand() *cobra.Command {
	var (
		organization string
		name         string // team name
	)

	cmd := &cobra.Command{
		Use:           "del-user [username]",
		Short:         "Delete user from team",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.GetTeam(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			if err := a.RemoveTeamMembership(cmd.Context(), args[0], team.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully removed %s from %s\n", args[0], name)
			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")
	cmd.Flags().StringVar(&name, "team", "", "Team in which to perform action")
	cmd.MarkFlagRequired("team")

	return cmd
}
