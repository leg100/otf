package auth

import (
	"time"

	"github.com/leg100/otf"
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

	NewTeamOptions struct {
		Name         string `schema:"team_name,required"`
		Organization string `schema:"organization_name,required"`
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

func NewTeam(opts NewTeamOptions) *Team {
	team := Team{
		ID:           otf.NewID("team"),
		Name:         opts.Name,
		CreatedAt:    otf.CurrentTimestamp(),
		Organization: opts.Organization,
	}
	return &team
}

func (u *Team) String() string                         { return u.Name }
func (u *Team) OrganizationAccess() OrganizationAccess { return u.Access }

func (u *Team) IsOwners() bool {
	return u.Name == "owners"
}

func (u *Team) Update(opts UpdateTeamOptions) error {
	u.Access.ManageWorkspaces = opts.ManageWorkspaces
	u.Access.ManageVCS = opts.ManageVCS
	u.Access.ManageRegistry = opts.ManageRegistry
	return nil
}
