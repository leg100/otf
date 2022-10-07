package otf

import (
	"context"
	"time"
)

// Team is a group of users sharing a level of authorization.
type Team struct {
	// ID uniquely identifies team
	id        string
	createdAt time.Time
	name      string
	// A team belongs to an organization
	organization *Organization
}

func (u *Team) ID() string                  { return u.id }
func (u *Team) Name() string                { return u.name }
func (u *Team) CreatedAt() time.Time        { return u.createdAt }
func (u *Team) String() string              { return u.name }
func (u *Team) Organization() *Organization { return u.organization }

// TeamService provides methods to interact with team accounts and their
// sessions.
type TeamService interface {
	// CreateTeam creates a team with the given name belong to the named
	// organization.
	CreateTeam(ctx context.Context, name, organizationName string) (*Team, error)
	// EnsureCreatedTeam retrieves a team; if they don't exist they'll be
	// created.
	EnsureCreatedTeam(ctx context.Context, name, organizationName string) (*Team, error)
	// Get retrieves a team according to the spec.
	GetTeam(ctx context.Context, spec TeamSpec) (*Team, error)
}

// TeamStore is a persistence store for team accounts.
type TeamStore interface {
	CreateTeam(ctx context.Context, team *Team) error
	GetTeam(ctx context.Context, spec TeamSpec) (*Team, error)
	ListTeams(ctx context.Context) ([]*Team, error)
	DeleteTeam(ctx context.Context, spec TeamSpec) error
}

type TeamSpec struct {
	TeamID           *string
	Name             *string
	OrganizationName *string
}

func NewTeam(name string, org *Organization, opts ...NewTeamOption) *Team {
	team := Team{
		id:           NewID("team"),
		name:         name,
		createdAt:    CurrentTimestamp(),
		organization: org,
	}
	for _, o := range opts {
		o(&team)
	}
	return &team
}

type NewTeamOption func(*Team)
