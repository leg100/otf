package auth

import (
	"context"
	"time"

	"github.com/leg100/otf"
)

// Team is a group of users sharing a level of authorization.
type Team struct {
	id           string // ID uniquely identifies team
	createdAt    time.Time
	name         string
	organization string // name of team's organization
	access       OrganizationAccess
}

func (u *Team) ID() string                             { return u.id }
func (u *Team) Name() string                           { return u.name }
func (u *Team) TeamName() string                       { return u.name }
func (u *Team) CreatedAt() time.Time                   { return u.createdAt }
func (u *Team) String() string                         { return u.name }
func (u *Team) Organization() string                   { return u.organization }
func (u *Team) OrganizationAccess() OrganizationAccess { return u.access }

func (u *Team) isOwners() bool {
	return u.name == "owners"
}

func (u *Team) Update(opts UpdateTeamOptions) error {
	u.access.ManageWorkspaces = opts.ManageWorkspaces
	u.access.ManageVCS = opts.ManageVCS
	u.access.ManageRegistry = opts.ManageRegistry
	return nil
}

func (u *Team) toValue() otf.Team {
	return otf.Team{
		ID:           u.id,
		Name:         u.name,
		Organization: u.organization,
		IsOwners:     u.isOwners(),
	}
}

// TeamService provides methods to interact with team accounts and their
// sessions.
type TeamService interface {
	// CreateTeam creates a team with the given name belong to the named
	// organization.
	CreateTeam(ctx context.Context, opts createTeamOptions) (*Team, error)
	UpdateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error)
	// Get retrieves a team with the given ID
	GetTeam(ctx context.Context, teamID string) (*Team, error)
	// ListTeams lists teams in an organization.
	ListTeams(ctx context.Context, organization string) ([]otf.Team, error)
	// ListTeamMembers lists users that are members of the given team
	ListTeamMembers(ctx context.Context, teamID string) ([]otf.User, error)
}

// TeamStore is a persistence store for team accounts.
type TeamStore interface {
	CreateTeam(ctx context.Context, team *Team) error
	UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error)
	GetTeam(ctx context.Context, name, organization string) (*Team, error)
	GetTeamByID(ctx context.Context, teamID string) (*Team, error)
	DeleteTeam(ctx context.Context, teamID string) error
	ListTeams(ctx context.Context, organization string) ([]*Team, error)
	// ListTeamMembers lists users that are members of the given team
	ListTeamMembers(ctx context.Context, teamID string) ([]otf.User, error)
}

// OrganizationAccess defines a team's organization access.
type OrganizationAccess struct {
	ManageWorkspaces bool `schema:"manage_workspaces"` // admin access on all workspaces
	ManageVCS        bool `schema:"manage_vcs"`        // manage VCS providers
	ManageRegistry   bool `schema:"manage_registry"`   // manage module and provider registry
}

type createTeamOptions struct {
	Name         string `schema:"team_name,required"`
	Organization string `schema:"organization_name,required"`
}

type UpdateTeamOptions struct {
	OrganizationAccess
}

func newTeam(opts createTeamOptions) *Team {
	team := Team{
		id:           otf.NewID("team"),
		name:         opts.Name,
		createdAt:    otf.CurrentTimestamp(),
		organization: opts.Organization,
	}
	return &team
}
