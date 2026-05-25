package cli

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	teamapi "github.com/leg100/otf/internal/team/api"
	"github.com/leg100/otf/internal/user"
	userapi "github.com/leg100/otf/internal/user/api"
	"github.com/spf13/cobra"
)

type membershipCLI struct {
	client membershipClient
}

type membershipClient interface {
	AddTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []user.Username) error
	RemoveTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []user.Username) error
	GetTeam(ctx context.Context, org organization.Name, name string) (*team.Team, error)
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
			cli.client = struct {
				*userapi.UserClient
				*teamapi.TeamClient
			}{
				UserClient: &userapi.Client{Client: apiclient},
				TeamClient: &teamapi.Client{Client: apiclient},
			}
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
			team, err := a.client.GetTeam(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			usernames := make([]user.Username, len(args))
			for i, arg := range args {
				username, err := user.NewUsername(arg)
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
			team, err := a.client.GetTeam(cmd.Context(), organization, name)
			if err != nil {
				return err
			}
			usernames := make([]user.Username, len(args))
			for i, arg := range args {
				username, err := user.NewUsername(arg)
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
