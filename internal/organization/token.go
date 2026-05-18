package organization

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

type (
	// OrganizationToken provides information about an API token for an organization
	OrganizationToken struct {
		ID        resource.TfeID `db:"organization_token_id"`
		CreatedAt time.Time      `db:"created_at"`
		// Token belongs to an organization
		Organization Name `db:"organization_name"`
		// Optional expiry.
		Expiry *time.Time
	}

	// CreateOrganizationTokenOptions are options for creating an organization token via the service
	// endpoint
	CreateOrganizationTokenOptions struct {
		Organization Name `schema:"organization_name,required"`
		Expiry       *time.Time
	}

	// tokenFactory constructs organization tokens
	tokenFactory struct {
		tokens *tokens.Service
	}
)

func (f *tokenFactory) NewOrganizationToken(opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	ot := OrganizationToken{
		ID:           resource.NewTfeID(resource.OrganizationTokenKind),
		CreatedAt:    internal.CurrentTimestamp(nil),
		Organization: opts.Organization,
		Expiry:       opts.Expiry,
	}
	token, err := f.tokens.NewToken(ot.ID, opts.Expiry)
	if err != nil {
		return nil, nil, err
	}
	return &ot, token, nil
}

func (u *OrganizationToken) String() string {
	return u.ID.String()
}

func (u *OrganizationToken) CanAccess(action resource.Action, kind resource.Kind, req authz.Request) bool {
	if req.ID == resource.SiteID {
		// Organization token cannot take action on site-level resources
		return false
	}
	if req.Organization() != nil && req.Organization().String() != u.Organization.String() {
		// Organization token cannot take action on other organizations
		return false
	}
	// can perform most actions in an organization, so it is easier to first refuse
	// access to those actions it CANNOT perform.
	switch kind {
	case resource.RunKind:
		switch action {
		case resource.EnqueuePlan, resource.Apply, resource.Cancel, resource.ForceCancel, resource.Discard, resource.Create, resource.Get, resource.List:
			return false
		}
	case resource.StateVersionKind:
		switch action {
		case resource.Create, resource.Rollback:
			return false
		}
	case resource.ChunkKind:
		switch action {
		case resource.Get, resource.Upload:
			return false
		}
	}
	return true
}
