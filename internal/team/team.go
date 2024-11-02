// Package team manages teams, which are groups of users with shared privileges.
package team

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

type (
	// Team is a group of users sharing a level of authorization.
	Team struct {
		ID           string    `jsonapi:"primary,teams"`
		CreatedAt    time.Time `jsonapi:"attribute" json:"created-at"`
		Name         string    `jsonapi:"attribute" json:"name"`
		Organization string    `jsonapi:"attribute" json:"organization"`

		Access OrganizationAccess

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Visibility string
		SSOTeamID  *string
	}

	CreateTeamOptions struct {
		// Name of team to create
		Name *string `json:"name" schema:"name,required"`

		OrganizationAccessOptions

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		SSOTeamID  *string
		Visibility *string
	}

	UpdateTeamOptions struct {
		Name *string

		OrganizationAccessOptions

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		SSOTeamID  *string
		Visibility *string
	}

	// OrganizationAccess defines a team's organization access.
	OrganizationAccess struct {
		ManageWorkspaces bool // admin access on all workspaces
		ManageVCS        bool // manage VCS providers
		ManageModules    bool // manage module registry

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		ManageProviders       bool
		ManagePolicies        bool
		ManagePolicyOverrides bool
	}

	// OrganizationAccessOptions defines access to be granted upon team creation
	// or to grant/rescind to/from an existing team.
	OrganizationAccessOptions struct {
		ManageWorkspaces *bool `schema:"manage_workspaces"`
		ManageVCS        *bool `schema:"manage_vcs"`
		ManageModules    *bool `schema:"manage_modules"`

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		ManageProviders       *bool
		ManagePolicies        *bool
		ManagePolicyOverrides *bool
	}
)

func newTeam(organization string, opts CreateTeamOptions) (*Team, error) {
	// required parameters
	if opts.Name == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "name"}
	}
	// default parameters
	if opts.Visibility == nil {
		opts.Visibility = internal.String("secret")
	}

	team := &Team{
		ID:           resource.NewID("team"),
		Name:         *opts.Name,
		CreatedAt:    internal.CurrentTimestamp(nil),
		Organization: organization,
		SSOTeamID:    opts.SSOTeamID,
		Visibility:   *opts.Visibility,
	}
	if opts.ManageWorkspaces != nil {
		team.Access.ManageWorkspaces = *opts.ManageWorkspaces
	}
	if opts.ManageVCS != nil {
		team.Access.ManageVCS = *opts.ManageVCS
	}
	if opts.ManageModules != nil {
		team.Access.ManageModules = *opts.ManageModules
	}
	if opts.ManageProviders != nil {
		team.Access.ManageProviders = *opts.ManageProviders
	}
	if opts.ManagePolicies != nil {
		team.Access.ManagePolicies = *opts.ManagePolicies
	}
	if opts.ManagePolicyOverrides != nil {
		team.Access.ManagePolicyOverrides = *opts.ManagePolicyOverrides
	}
	return team, nil
}

func (t *Team) String() string                         { return t.Name }
func (t *Team) OrganizationAccess() OrganizationAccess { return t.Access }

func (t *Team) IsSiteAdmin() bool { return false }

func (t *Team) IsOwners() bool {
	return t.Name == "owners"
}

func (t *Team) IsOwner(organization string) bool {
	return t.Organization == organization && t.IsOwners()
}

func (t *Team) CanAccessSite(action rbac.Action) bool {
	return false
}

func (t *Team) CanAccessTeam(action rbac.Action, id resource.ID) bool {
	// team can access self
	return t.ID == id
}

func (t *Team) CanAccessOrganization(action rbac.Action, org string) bool {
	if t.Organization == org {
		if t.IsOwners() {
			// owner team can perform all actions on organization
			return true
		}
		if rbac.OrganizationMinPermissions.IsAllowed(action) {
			return true
		}
		if t.Access.ManageWorkspaces {
			if rbac.WorkspaceManagerRole.IsAllowed(action) {
				return true
			}
		}
		if t.Access.ManageVCS {
			if rbac.VCSManagerRole.IsAllowed(action) {
				return true
			}
		}
		if t.Access.ManageModules {
			if rbac.VCSManagerRole.IsAllowed(action) {
				return true
			}
		}
	}
	return false
}

func (t *Team) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// coarser-grained organization perms take precedence.
	if t.CanAccessOrganization(action, policy.Organization) {
		return true
	}
	// fallback to checking finer-grained workspace perms
	if t.Organization != policy.Organization {
		return false
	}
	for _, perm := range policy.Permissions {
		if t.ID == perm.TeamID {
			return perm.Role.IsAllowed(action)
		}
	}
	return false
}

func (t *Team) Organizations() []string {
	return []string{t.Organization}
}

func (t *Team) Update(opts UpdateTeamOptions) error {
	if opts.Name != nil {
		t.Name = *opts.Name
	}
	if opts.SSOTeamID != nil {
		t.SSOTeamID = opts.SSOTeamID
	}
	if opts.Visibility != nil {
		t.Visibility = *opts.Visibility
	}
	if opts.ManageWorkspaces != nil {
		t.Access.ManageWorkspaces = *opts.ManageWorkspaces
	}
	if opts.ManageVCS != nil {
		t.Access.ManageVCS = *opts.ManageVCS
	}
	if opts.ManageModules != nil {
		t.Access.ManageModules = *opts.ManageModules
	}
	if opts.ManageProviders != nil {
		t.Access.ManageProviders = *opts.ManageProviders
	}
	if opts.ManagePolicies != nil {
		t.Access.ManagePolicies = *opts.ManagePolicies
	}
	if opts.ManagePolicyOverrides != nil {
		t.Access.ManagePolicyOverrides = *opts.ManagePolicyOverrides
	}
	return nil
}
