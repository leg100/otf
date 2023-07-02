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

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		Visibility string
		SSOTeamID  *string
	}

	CreateTeamOptions struct {
		// Name of team to create
		Name *string `schema:"name,required"`

		OrganizationAccessOptions

		// TFE fields that OTF does not support but persists merely to pass the
		// go-tfe integration tests
		SSOTeamID  *string
		Visibility *string

		// Database transaction within which to create team. Optional.
		Tx internal.DB
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
		return nil, &internal.MissingParameterError{Parameter: "name"}
	}
	// default parameters
	if opts.Visibility == nil {
		opts.Visibility = internal.String("secret")
	}

	team := &Team{
		ID:           internal.NewID("team"),
		Name:         *opts.Name,
		CreatedAt:    internal.CurrentTimestamp(),
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

func (t *Team) IsOwners() bool {
	return t.Name == "owners"
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
