package auth

import (
	"time"

	"github.com/leg100/otf/internal"
)

type (
	// Team is a group of users sharing a level of authorization.
	Team struct {
		ID           string
		CreatedAt    time.Time
		Name         string
		Organization string

		Access OrganizationAccess
	}

	CreateTeamOptions struct {
		// Name of team to create
		Name string `schema:"name,required"`
		// Organization in which to creat team
		Organization string `schema:"organization_name,required"`
		// Database transaction within which to create team. Optional.
		Tx internal.DB
	}

	// OrganizationAccess defines a team's organization access.
	OrganizationAccess struct {
		ManageWorkspaces bool `schema:"manage_workspaces"` // admin access on all workspaces
		ManageVCS        bool `schema:"manage_vcs"`        // manage VCS providers
		ManageRegistry   bool `schema:"manage_registry"`   // manage module and provider registry
	}

	UpdateTeamOptions struct {
		OrganizationAccess
	}
)

func NewTeam(opts CreateTeamOptions) *Team {
	team := Team{
		ID:           internal.NewID("team"),
		Name:         opts.Name,
		CreatedAt:    internal.CurrentTimestamp(),
		Organization: opts.Organization,
	}
	return &team
}

func (t *Team) String() string                         { return t.Name }
func (t *Team) OrganizationAccess() OrganizationAccess { return t.Access }

func (t *Team) IsOwners() bool {
	return t.Name == "owners"
}

func (t *Team) Update(opts UpdateTeamOptions) error {
	t.Access.ManageWorkspaces = opts.ManageWorkspaces
	t.Access.ManageVCS = opts.ManageVCS
	t.Access.ManageRegistry = opts.ManageRegistry
	return nil
}
