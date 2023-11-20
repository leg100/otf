package user

import (
	"fmt"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/team"
	"github.com/spf13/cobra"
)

type userCLI struct {
	UserService
}

func NewUserCommand(api *otfapi.Client) *cobra.Command {
	cli := &userCLI{}
	cmd := &cobra.Command{
		Use:   "users",
		Short: "User account management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.UserService = &client{Client: api}
			return nil
		},
	}

	cmd.AddCommand(cli.userNewCommand())
	cmd.AddCommand(cli.userDeleteCommand())

	return cmd
}

func (a *userCLI) userNewCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "new [username]",
		Short:         "Create a new user account",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := a.CreateUser(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created user %s\n", user.Username)
			return nil
		},
	}
}

func (a *userCLI) userDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "delete [username]",
		Short:         "Delete a user account",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.DeleteUser(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted user %s\n", args[0])
			return nil
		},
	}
}

type membershipCLI struct {
	UserService
	team.TeamService
}

func NewTeamMembershipCommand(apiclient *otfapi.Client) *cobra.Command {
	cli := &membershipCLI{}
	cmd := &cobra.Command{
		Use:   "team-membership",
		Short: "Team membership management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.UserService = &client{Client: apiclient}
			cli.TeamService = &team.Client{Client: apiclient}
			return nil
		},
	}

	cmd.AddCommand(cli.addTeamMembershipCommand())
	cmd.AddCommand(cli.deleteTeamMembershipCommand())

	return cmd
}

func (a *membershipCLI) addTeamMembershipCommand() *cobra.Command {
	var (
		organization string
		name         string // team name
	)

	cmd := &cobra.Command{
		Use:           "add-users [username|...]",
		Short:         "Add users to team",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.GetTeam(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			if err := a.AddTeamMembership(cmd.Context(), team.ID, args); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully added %s to %s\n", args, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")
	cmd.Flags().StringVar(&name, "team", "", "Team in which to perform action")
	cmd.MarkFlagRequired("team")

	return cmd
}

func (a *membershipCLI) deleteTeamMembershipCommand() *cobra.Command {
	var (
		organization string
		name         string // team name
	)

	cmd := &cobra.Command{
		Use:           "del-users [username|...]",
		Short:         "Delete users from team",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.GetTeam(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			if err := a.RemoveTeamMembership(cmd.Context(), team.ID, args); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully removed %s from %s\n", args, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&organization, "organization", "", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")
	cmd.Flags().StringVar(&name, "team", "", "Team in which to perform action")
	cmd.MarkFlagRequired("team")

	return cmd
}
