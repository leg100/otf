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
		resource.ID

		CreatedAt time.Time
		// Token belongs to an organization
		Organization string
		// Optional expiry.
		Expiry *time.Time
	}

	// CreateOrganizationTokenOptions are options for creating an organization token via the service
	// endpoint
	CreateOrganizationTokenOptions struct {
		Organization string `schema:"organization_name,required"`
		Expiry       *time.Time
	}

	// tokenFactory constructs organization tokens
	tokenFactory struct {
		tokens *tokens.Service
	}
)

func (f *tokenFactory) NewOrganizationToken(opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	ot := OrganizationToken{
		ID:           resource.NewID(resource.OrganizationTokenKind),
		CreatedAt:    internal.CurrentTimestamp(nil),
		Organization: opts.Organization,
		Expiry:       opts.Expiry,
	}
	var newTokenOptions []tokens.NewTokenOption
	if opts.Expiry != nil {
		newTokenOptions = append(newTokenOptions, tokens.WithExpiry(*opts.Expiry))
	}
	token, err := f.tokens.NewToken(ot.ID, newTokenOptions...)
	if err != nil {
		return nil, nil, err
	}
	return &ot, token, nil
}

func (u *OrganizationToken) CanAccess(action authz.Action, req *authz.AccessRequest) bool {
	if req == nil {
		// Organization token cannot take action on site-level resources
		return false
	}
	if req.Organization != u.Organization {
		// Organization token cannot take action on other organizations
		return false
	}
	// can perform most actions in an organization, so it is easier to first refuse
	// access to those actions it CANNOT perform.
	switch action {
	case authz.GetRunAction, authz.ListRunsAction, authz.ApplyRunAction, authz.CreateRunAction, authz.DiscardRunAction, authz.CancelRunAction, authz.ForceCancelRunAction, authz.EnqueuePlanAction, authz.PutChunkAction, authz.TailLogsAction, authz.CreateStateVersionAction, authz.RollbackStateVersionAction:
		return false
	}
	return true
}

func (u *OrganizationToken) Organizations() []string {
	return []string{u.Organization}
}
