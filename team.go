package otf

import "context"

type Team interface {
	ID() string
	Name() string
	Organization() string
	IsOwners() bool
}

type TeamService interface {
	EnsureCreatedTeam(ctx context.Context, opts CreateTeamOptions) (Team, error)
	// Get retrieves a team with the given ID
	GetTeam(ctx context.Context, teamID string) (Team, error)
}

type CreateTeamOptions struct {
	Name         string `schema:"team_name,required"`
	Organization string `schema:"organization_name,required"`
}
