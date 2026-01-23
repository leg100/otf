package user

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/spf13/cobra"
)

type userCLI struct {
	client userCLIClient
}

type userCLIClient interface {
	Create(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
	Delete(ctx context.Context, username Username) error
}

func NewUserCommand(apiClient *otfhttp.Client) *cobra.Command {
	cli := &userCLI{}
	cmd := &cobra.Command{
		Use:   "users",
		Short: "User account management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.client = &Client{Client: apiClient}
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
			user, err := a.client.Create(cmd.Context(), args[0])
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
			if err := a.client.Delete(cmd.Context(), Username{name: args[0]}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted user %s\n", args[0])
			return nil
		},
	}
}

type membershipCLI struct {
	client membershipCLIClient
	teams  teamsCLIClient
}

type membershipCLIClient interface {
	AddTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []Username) error
	RemoveTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []Username) error
}

type teamsCLIClient interface {
	Get(ctx context.Context, org organization.Name, name string) (*team.Team, error)
}

func NewTeamMembershipCommand(apiclient *otfhttp.Client) *cobra.Command {
	cli := &membershipCLI{}
	cmd := &cobra.Command{
		Use:   "team-membership",
		Short: "Team membership management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.client = &Client{Client: apiclient}
			cli.teams = &team.Client{Client: apiclient}
			return nil
		},
	}

	cmd.AddCommand(cli.addTeamMembershipCommand())
	cmd.AddCommand(cli.deleteTeamMembershipCommand())

	return cmd
}

func (a *membershipCLI) addTeamMembershipCommand() *cobra.Command {
	var (
		organization organization.Name
		name         string // team name
	)

	cmd := &cobra.Command{
		Use:           "add-users [username|...]",
		Short:         "Add users to team",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.teams.Get(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			usernames := make([]Username, len(args))
			for i, arg := range args {
				username, err := NewUsername(arg)
				if err != nil {
					return fmt.Errorf("invalid username: %w", err)
				}
				usernames[i] = username
			}
			if err := a.client.AddTeamMembership(cmd.Context(), team.ID, usernames); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully added %s to %s\n", args, name)
			return nil
		},
	}

	cmd.Flags().Var(&organization, "organization", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")
	cmd.Flags().StringVar(&name, "team", "", "Team in which to perform action")
	cmd.MarkFlagRequired("team")

	return cmd
}

func (a *membershipCLI) deleteTeamMembershipCommand() *cobra.Command {
	var (
		organization organization.Name
		name         string // team name
	)

	cmd := &cobra.Command{
		Use:           "del-users [username|...]",
		Short:         "Delete users from team",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.teams.Get(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			usernames := make([]Username, len(args))
			for i, arg := range args {
				username, err := NewUsername(arg)
				if err != nil {
					return fmt.Errorf("invalid username: %w", err)
				}
				usernames[i] = username
			}
			if err := a.client.RemoveTeamMembership(cmd.Context(), team.ID, usernames); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully removed %s from %s\n", args, name)
			return nil
		},
	}

	cmd.Flags().Var(&organization, "organization", "OTF organization in which to perform action")
	cmd.MarkFlagRequired("organization")
	cmd.Flags().StringVar(&name, "team", "", "Team in which to perform action")
	cmd.MarkFlagRequired("team")

	return cmd
}
