package otf

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidTeamSpec = errors.New("invalid team spec options")

// Team is a group of users sharing a level of authorization.
type Team struct {
	// ID uniquely identifies team
	id        string
	createdAt time.Time
	name      string

	// TODO: remove
	organizationName string

	// A team belongs to an organization
	organizationID string

	access OrganizationAccess
}

func (u *Team) ID() string                             { return u.id }
func (u *Team) Name() string                           { return u.name }
func (u *Team) TeamName() string                       { return u.name }
func (u *Team) CreatedAt() time.Time                   { return u.createdAt }
func (u *Team) String() string                         { return u.name }
func (u *Team) OrganizationID() string                 { return u.organizationID }
func (u *Team) OrganizationAccess() OrganizationAccess { return u.access }
func (u *Team) OrganizationName() string               { return u.organizationName }

func (u *Team) IsOwners() bool {
	return u.name == "owners"
}

func (u *Team) Update(opts TeamUpdateOptions) error {
	u.access.ManageWorkspaces = opts.ManageWorkspaces
	return nil
}

// TeamService provides methods to interact with team accounts and their
// sessions.
type TeamService interface {
	// CreateTeam creates a team with the given name belong to the named
	// organization.
	CreateTeam(ctx context.Context, name, organizationName string) (*Team, error)
	UpdateTeam(ctx context.Context, spec TeamSpec, opts TeamUpdateOptions) (*Team, error)
	// EnsureCreatedTeam retrieves a team; if they don't exist they'll be
	// created.
	EnsureCreatedTeam(ctx context.Context, name, organizationName string) (*Team, error)
	// Get retrieves a team according to the spec.
	GetTeam(ctx context.Context, spec TeamSpec) (*Team, error)
	// ListTeams lists teams in an organization.
	ListTeams(ctx context.Context, organizationName string) ([]*Team, error)
}

// TeamStore is a persistence store for team accounts.
type TeamStore interface {
	CreateTeam(ctx context.Context, team *Team) error
	UpdateTeam(ctx context.Context, spec TeamSpec, fn func(*Team) error) (*Team, error)
	GetTeam(ctx context.Context, spec TeamSpec) (*Team, error)
	DeleteTeam(ctx context.Context, spec TeamSpec) error
	ListTeams(ctx context.Context, organizationName string) ([]*Team, error)
}

type TeamSpec struct {
	ID               *string
	Name             *string
	OrganizationName *string
}

// KeyValue returns the team spec in key-value form. Useful for logging
// purposes.
func (spec *TeamSpec) KeyValue() []any {
	if spec.ID != nil {
		return []interface{}{"id", *spec.ID}
	}
	if spec.Name != nil && spec.OrganizationName != nil {
		return []interface{}{"name", *spec.Name, "organization", *spec.OrganizationName}
	}
	return []any{}
}

// OrganizationAccess defines a team's organization access.
type OrganizationAccess struct {
	ManageWorkspaces bool
}

type TeamUpdateOptions struct {
	ManageWorkspaces bool `schema:"manage_workspaces"`
}

func NewTeam(name string, org *Organization, opts ...NewTeamOption) *Team {
	team := Team{
		id:               NewID("team"),
		name:             name,
		createdAt:        CurrentTimestamp(),
		organizationName: org.Name(),
		organizationID:   org.ID(),
	}
	for _, o := range opts {
		o(&team)
	}
	return &team
}

type NewTeamOption func(*Team)
