// Package agent contains code related to agents
package agent

import (
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

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

func (p *Pool) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", p.ID),
		slog.String("name", p.Name),
		slog.String("organization", p.Organization),
		slog.Bool("organization_scoped", p.OrganizationScoped),
		slog.Any("allowed_workspaces", p.AllowedWorkspaces),
	)
}
