// Package team manages teams, which are groups of users with shared privileges.
package team

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

type (
	// Team is a group of users sharing a level of authorization.
	Team struct {
		ID        resource.TfeID `jsonapi:"primary,teams" db:"team_id"`
		Name      string         `jsonapi:"attribute" json:"name" db:"name"`
		CreatedAt time.Time      `jsonapi:"attribute" json:"created-at" db:"created_at"`

		ManageWorkspaces bool `db:"permission_manage_workspaces"` // admin access on all workspaces
		ManageVCS        bool `db:"permission_manage_vcs"`        // manage VCS providers
		ManageModules    bool `db:"permission_manage_modules"`    // manage module registry

		Organization organization.Name `jsonapi:"attribute" json:"organization" db:"organization_name"`

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		SSOTeamID  *string `db:"sso_team_id"`
		Visibility string

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		ManagePolicies        bool `db:"permission_manage_policies"`
		ManagePolicyOverrides bool `db:"permission_manage_policy_overrides"`
		ManageProviders       bool `db:"permission_manage_providers"`
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

	ListOptions struct {
		resource.PageOptions
		Organization organization.Name `schema:"organization_name"`
	}

	// OrganizationAccessOptions defines access to be granted upon team creation
	// or to grant/rescind to/from an existing team.
	OrganizationAccessOptions struct {
		ManageWorkspaces *bool
		ManageVCS        *bool
		ManageModules    *bool

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		ManageProviders       *bool
		ManagePolicies        *bool
		ManagePolicyOverrides *bool
	}
)

func newTeam(organization organization.Name, opts CreateTeamOptions) (*Team, error) {
	// required parameters
	if opts.Name == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "name"}
	}
	// default parameters
	if opts.Visibility == nil {
		opts.Visibility = new("secret")
	}

	team := &Team{
		ID:           resource.NewTfeID("team"),
		Name:         *opts.Name,
		CreatedAt:    internal.CurrentTimestamp(nil),
		Organization: organization,
		SSOTeamID:    opts.SSOTeamID,
		Visibility:   *opts.Visibility,
	}
	if opts.ManageWorkspaces != nil {
		team.ManageWorkspaces = *opts.ManageWorkspaces
	}
	if opts.ManageVCS != nil {
		team.ManageVCS = *opts.ManageVCS
	}
	if opts.ManageModules != nil {
		team.ManageModules = *opts.ManageModules
	}
	if opts.ManageProviders != nil {
		team.ManageProviders = *opts.ManageProviders
	}
	if opts.ManagePolicies != nil {
		team.ManagePolicies = *opts.ManagePolicies
	}
	if opts.ManagePolicyOverrides != nil {
		team.ManagePolicyOverrides = *opts.ManagePolicyOverrides
	}
	return team, nil
}

func (t *Team) String() string { return t.Name }

func (t *Team) IsOwners() bool {
	return t.Name == "owners"
}

func (t *Team) IsOwner(organization resource.ID) bool {
	return t.Organization == organization && t.IsOwners()
}

func (t *Team) CanAccess(action authz.Action, req authz.Request) bool {
	if req.ID == resource.SiteID {
		// Deny all site-level access
		return false
	}
	if req.Organization() != nil && req.Organization().String() != t.Organization.String() {
		// Deny access to other organizations
		return false
	}
	if t.IsOwners() {
		// owner team can perform all actions on organization
		return true
	}
	if authz.OrganizationMinPermissions.IsAllowed(action) {
		return true
	}
	if t.ManageWorkspaces {
		if authz.WorkspaceManagerRole.IsAllowed(action) {
			return true
		}
	}
	if t.ManageVCS {
		if authz.VCSManagerRole.IsAllowed(action) {
			return true
		}
	}
	if t.ManageModules {
		if authz.RegistryManagerRole.IsAllowed(action) {
			return true
		}
	}
	if req.ID != nil && req.ID.Kind() == resource.TeamKind {
		// team can access self
		return t.ID == req.ID
	}
	if req.WorkspacePolicy != nil {
		return req.WorkspacePolicy.Check(t.ID, action)
	}
	return false
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
		t.ManageWorkspaces = *opts.ManageWorkspaces
	}
	if opts.ManageVCS != nil {
		t.ManageVCS = *opts.ManageVCS
	}
	if opts.ManageModules != nil {
		t.ManageModules = *opts.ManageModules
	}
	if opts.ManageProviders != nil {
		t.ManageProviders = *opts.ManageProviders
	}
	if opts.ManagePolicies != nil {
		t.ManagePolicies = *opts.ManagePolicies
	}
	if opts.ManagePolicyOverrides != nil {
		t.ManagePolicyOverrides = *opts.ManagePolicyOverrides
	}
	return nil
}
