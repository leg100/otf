package runner

import (
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

var (
	ErrCannotDeletePoolReferencedByWorkspaces = errors.New("agent pool is still being used by workspaces in your organization. You must switch your workspaces to a different agent pool or execution mode before you can delete this agent pool")
	ErrWorkspaceNotAllowedToUsePool           = errors.New("access to this agent pool is not allowed - you must explictly grant access to the workspace first")
	ErrPoolAssignedWorkspacesNotAllowed       = errors.New("workspaces assigned to the pool have not been granted access to the pool")
)

type (
	// Pool is a group of remote runners sharing one or more tokens, assigned to
	// an organization or particular workspaces within the organization.
	Pool struct {
		ID        resource.TfeID `json:"agent_pool_id" db:"agent_pool_id"`
		Name      string
		CreatedAt time.Time `json:"created_at" db:"created_at"`
		// Pool belongs to an organization with this name.
		Organization organization.Name `json:"organization_name" db:"organization_name"`
		// Whether pool of agents is accessible to all workspaces in organization
		// (true) or only those specified in AllowedWorkspaces (false).
		OrganizationScoped bool `json:"organization_scoped" db:"organization_scoped"`
		// IDs of workspaces allowed to access pool. Ignored if OrganizationScoped
		// is true.
		AllowedWorkspaces []resource.TfeID `json:"allowed_workspace_ids" db:"allowed_workspace_ids"`
		// IDs of workspaces assigned to the pool. Note: this is a subset of
		// AllowedWorkspaces.
		AssignedWorkspaces []resource.TfeID `json:"workspace_ids" db:"workspace_ids"`
	}

	CreateAgentPoolOptions struct {
		Name string `schema:"name,required"`
		// name of org
		Organization organization.Name `schema:"organization_name,required"`
		// defaults to true
		OrganizationScoped *bool
		// IDs of workspaces allowed to access the pool.
		AllowedWorkspaces []resource.TfeID
	}

	UpdatePoolOptions struct {
		Name               *string
		OrganizationScoped *bool `schema:"organization_scoped"`
		// IDs of workspaces allowed to access the pool.
		AllowedWorkspaces []resource.TfeID `schema:"allowed_workspaces"`
		// IDs of workspaces assigned to the pool. Note: this is a subset of
		// AssignedWorkspaces.
		AssignedWorkspaces []resource.TfeID `schema:"assigned_workspaces"`
	}

	ListPoolOptions struct {
		// Filter pools by those with this substring in their name. Optional.
		NameSubstring *string
		// Filter pools to those accessible to the named workspace. Optional.
		AllowedWorkspaceName *string
		// Filter pools to those accessible to the workspace with the given ID. Optional.
		AllowedWorkspaceID *resource.TfeID
	}
)

// newPool constructs a new agent pool. Note: a new pool has a list of allowed
// workspaces but not yet a list of assigned workspaces.
func newPool(opts CreateAgentPoolOptions) (*Pool, error) {
	if opts.Name == "" {
		return nil, errors.New("name must not be empty")
	}
	pool := &Pool{
		ID:                 resource.NewTfeID("apool"),
		CreatedAt:          internal.CurrentTimestamp(nil),
		Name:               opts.Name,
		Organization:       opts.Organization,
		OrganizationScoped: true,
		AllowedWorkspaces:  opts.AllowedWorkspaces,
	}
	if opts.OrganizationScoped != nil {
		pool.OrganizationScoped = *opts.OrganizationScoped
	}
	return pool, nil
}

func (p *Pool) update(opts UpdatePoolOptions) error {
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
	// if not organization scoped then each assigned workspace must also be
	// allowed.
	if !p.OrganizationScoped {
		for _, assigned := range p.AssignedWorkspaces {
			if !slices.Contains(p.AllowedWorkspaces, assigned) {
				return ErrPoolAssignedWorkspacesNotAllowed
			}
		}
	}
	return nil
}

func (p *Pool) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", p.ID.String()),
		slog.String("name", p.Name),
		slog.Any("organization", p.Organization),
		slog.Bool("organization_scoped", p.OrganizationScoped),
		slog.Any("workspaces", p.AssignedWorkspaces),
		slog.Any("allowed_workspaces", p.AllowedWorkspaces),
	)
}
