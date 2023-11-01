// Package agent contains code related to agents
package agent

import "github.com/leg100/otf/internal/resource"

type (
	// Pool is a group of non-server agents sharing one or more tokens, assigned to
	// an organization or particular workspaces within the organization.
	Pool struct {
		// Unique system-wide ID
		ID   string
		Name string
		// Pool belongs to an organization with this name.
		Organization string
		// Whether pool of agents is accessible to all workspaces in organization
		// (true) or only those specified in AllowedWorkspaces (false).
		OrganizationScoped bool
		// IDs of workspaces allowed to access pool. Ignored if OrganizationScoped
		// is true.
		AllowedWorkspaces []string
	}

	createPoolOptions struct {
		Name         *string
		Organization string
	}
)

func newPool(opts createPoolOptions) (*Pool, error) {
	if err := resource.ValidateName(opts.Name); err != nil {
		return nil, err
	}
	return &Pool{}, nil
}
