package team

import (
	"fmt"

	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"

	"github.com/spf13/cobra"
)

type teamCLI struct {
	TeamService
}

func NewTeamCommand(api *otfapi.Client) *cobra.Command {
	cli := &teamCLI{}
	cmd := &cobra.Command{
		Use:   "teams",
		Short: "Team management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.TeamService = &Client{Client: api}
			return nil
		},
	}
	cmd.AddCommand(cli.teamNewCommand())
	cmd.AddCommand(cli.teamDeleteCommand())

	return cmd
}

func (a *teamCLI) teamNewCommand() *cobra.Command {
	var organization string

	cmd := &cobra.Command{
		Use:           "new [name]",
		Short:         "Create a new team",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := a.CreateTeam(cmd.Context(), organization, CreateTeamOptions{
				Name: internal.String(args[0]),
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

func (a *teamCLI) teamDeleteCommand() *cobra.Command {
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
