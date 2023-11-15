// Package agent contains code related to agents
package agent

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

// An agent pool implements Subject (otf-agent authenticates as an agent pool using one
// of its tokens).
var _ internal.Subject = (*Pool)(nil)

type (
	// Pool is a group of non-server agents sharing one or more tokens, assigned to
	// an organization or particular workspaces within the organization.
	Pool struct {
		// Unique system-wide ID
		ID        string
		Name      string
		CreatedAt time.Time
		// Pool belongs to an organization with this name.
		Organization string
		// Whether pool of agents is accessible to all workspaces in organization
		// (true) or only those specified in AllowedWorkspaces (false).
		OrganizationScoped bool
		// IDs of workspaces currently configured to use this pool
		Workspaces []string
		// IDs of workspaces allowed to access pool. Ignored if OrganizationScoped
		// is true.
		AllowedWorkspaces []string
	}

	createPoolOptions struct {
		Name               *string
		Organization       string   // name of org
		OrganizationScoped *bool    // defaults to true
		AllowedWorkspaces  []string // IDs of workspaces
	}

	updatePoolOptions struct {
		Name               *string
		OrganizationScoped *bool
		AllowedWorkspaces  []string
	}

	listPoolOptions struct {
		// Filter by organization name. Optional.
		Organization *string
		// Filter pools by those with this substring in their name. Optional.
		NameSubstring *string
		// Filter pools to those accessible to the named workspace. Optional.
		AllowedWorkspaceName *string
	}
)

func newPool(opts createPoolOptions) (*Pool, error) {
	if err := resource.ValidateName(opts.Name); err != nil {
		return nil, err
	}
	pool := &Pool{
		ID:                 internal.NewID("apool"),
		CreatedAt:          internal.CurrentTimestamp(nil),
		Name:               *opts.Name,
		Organization:       opts.Organization,
		OrganizationScoped: true,
		AllowedWorkspaces:  opts.AllowedWorkspaces,
	}
	if opts.OrganizationScoped != nil {
		pool.OrganizationScoped = *opts.OrganizationScoped
	}
	return pool, nil
}

func (p *Pool) update(opts updatePoolOptions) error {
	if opts.Name != nil {
		if err := resource.ValidateName(opts.Name); err != nil {
			return err
		}
		p.Name = *opts.Name
	}
	if opts.OrganizationScoped != nil {
		p.OrganizationScoped = *opts.OrganizationScoped
	}
	if opts.AllowedWorkspaces != nil {
		p.AllowedWorkspaces = opts.AllowedWorkspaces
	}
	return nil
}

func (a *Pool) String() string      { return a.ID }
func (a *Pool) IsSiteAdmin() bool   { return true }
func (a *Pool) IsOwner(string) bool { return true }

func (a *Pool) Organizations() []string { return []string{a.Organization} }

func (*Pool) CanAccessSite(action rbac.Action) bool {
	// agent pool cannot carry out site-level actions
	return false
}

func (*Pool) CanAccessTeam(rbac.Action, string) bool {
	// agent pool cannot carry out team-level actions
	return false
}

func (a *Pool) CanAccessOrganization(action rbac.Action, name string) bool {
	// agent pool can access anything within its organization
	//
	// TODO: permit only those actions that an agent pool needs to carry out
	// (get agent jobs, etc).
	return a.Organization == name
}

func (a *Pool) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// agent pool can access anything within its organization
	//
	// TODO: permit only those actions that an agent pool needs to carry out
	// (get agent jobs, etc).
	return a.Organization == policy.Organization
}

// PoolFromContext retrieves an agent pool subject from a context
func PoolFromContext(ctx context.Context) (*Pool, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	pool, ok := subj.(*Pool)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent pool")
	}
	return pool, nil
}

func (p *Pool) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", p.ID),
		slog.String("name", p.Name),
		slog.String("organization", p.Organization),
		slog.Bool("organization_scoped", p.OrganizationScoped),
		slog.Any("workspaces", p.Workspaces),
		slog.Any("allowed_workspaces", p.AllowedWorkspaces),
	)
}
