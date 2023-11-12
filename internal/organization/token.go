package organization

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/tokens"
)

const OrganizationTokenKind tokens.Kind = "organization_token"

type (
	// OrganizationToken provides information about an API token for an organization
	OrganizationToken struct {
		ID        string
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
		tokens.TokensService
	}
)

func (f *tokenFactory) NewOrganizationToken(opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	ot := OrganizationToken{
		ID:           internal.NewID("ot"),
		CreatedAt:    internal.CurrentTimestamp(nil),
		Organization: opts.Organization,
		Expiry:       opts.Expiry,
	}
	token, err := f.NewToken(tokens.NewTokenOptions{
		Subject: ot.ID,
		Kind:    OrganizationTokenKind,
		Expiry:  opts.Expiry,
	})
	if err != nil {
		return nil, nil, err
	}
	return &ot, token, nil
}

func (u *OrganizationToken) CanAccessSite(action rbac.Action) bool {
	// only be used for organization-scoped resources.
	return false
}

func (u *OrganizationToken) CanAccessTeam(rbac.Action, string) bool {
	// only be used for organization-scoped resources.
	return false
}

func (u *OrganizationToken) CanAccessOrganization(action rbac.Action, org string) bool {
	if u.Organization != org {
		return false
	}
	// can perform most actions in an organization, so it is easier to first refuse
	// access to those actions it CANNOT perform.
	switch action {
	case rbac.GetRunAction, rbac.ListRunsAction, rbac.ApplyRunAction, rbac.CreateRunAction, rbac.DiscardRunAction, rbac.CancelRunAction, rbac.EnqueuePlanAction, rbac.StartPhaseAction, rbac.FinishPhaseAction, rbac.PutChunkAction, rbac.TailLogsAction, rbac.CreateStateVersionAction, rbac.RollbackStateVersionAction:
		return false
	}
	return true
}

func (u *OrganizationToken) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	return u.CanAccessOrganization(action, policy.Organization)
}

func (u *OrganizationToken) IsOwner(organization string) bool {
	// an owner would give perms to all actions in org whereas an org token
	// cannot perform certain actions, so org token is not an owner.
	return false
}

func (u *OrganizationToken) IsSiteAdmin() bool { return false }
func (u *OrganizationToken) String() string    { return u.ID }

func (u *OrganizationToken) Organizations() []string {
	return []string{u.Organization}
}
