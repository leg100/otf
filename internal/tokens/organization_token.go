package tokens

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	// OrganizationToken provides information about an API token for a user.
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

	// NewOrganizationTokenOptions are options for constructing a user token via the
	// constructor.
	NewOrganizationTokenOptions struct {
		CreateOrganizationTokenOptions
		Organization string
		key          jwk.Key
	}

	organizationTokenService interface {
		// CreateOrganizationToken creates a user token.
		CreateOrganizationToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error)
		// GetOrganizationToken gets the organization token. If a token does not
		// exist, then nil is returned without an error.
		GetOrganizationToken(ctx context.Context, organization string) (*OrganizationToken, error)
		// DeleteOrganizationToken deletes an organization token.
		DeleteOrganizationToken(ctx context.Context, organization string) error

		getOrganizationTokenByID(ctx context.Context, tokenID string) (*OrganizationToken, error)
	}
)

func NewOrganizationToken(opts NewOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	ot := OrganizationToken{
		ID:           internal.NewID("ot"),
		CreatedAt:    internal.CurrentTimestamp(),
		Organization: opts.Organization,
		Expiry:       opts.Expiry,
	}
	token, err := NewToken(NewTokenOptions{
		key:     opts.key,
		Subject: ot.ID,
		Kind:    organizationTokenKind,
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

func (u *OrganizationToken) CanAccessTeam(action rbac.Action, teamName string) bool {
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

// CreateOrganizationToken creates an organization token. If an organization
// token already exists it is replaced.
func (a *service) CreateOrganizationToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	_, err := a.organization.CanAccess(ctx, rbac.CreateOrganizationTokenAction, opts.Organization)
	if err != nil {
		return nil, nil, err
	}

	ot, token, err := NewOrganizationToken(NewOrganizationTokenOptions{
		CreateOrganizationTokenOptions: opts,
		Organization:                   opts.Organization,
		key:                            a.key,
	})
	if err != nil {
		a.Error(err, "constructing organization token", "organization", opts.Organization)
		return nil, nil, err
	}

	if err := a.db.upsertOrganizationToken(ctx, ot); err != nil {
		a.Error(err, "creating organization token", "organization", opts.Organization)
		return nil, nil, err
	}

	a.V(0).Info("created organization token", "organization", opts.Organization)

	return ot, token, nil
}

func (a *service) GetOrganizationToken(ctx context.Context, organization string) (*OrganizationToken, error) {
	return a.db.getOrganizationTokenByName(ctx, organization)
}

func (a *service) DeleteOrganizationToken(ctx context.Context, organization string) error {
	_, err := a.organization.CanAccess(ctx, rbac.CreateOrganizationTokenAction, organization)
	if err != nil {
		return err
	}

	if err := a.db.deleteOrganizationToken(ctx, organization); err != nil {
		a.Error(err, "deleting organization token", "organization", organization)
		return err
	}

	a.V(0).Info("deleted organization token", "organization", organization)

	return nil
}

func (a *service) getOrganizationTokenByID(ctx context.Context, tokenID string) (*OrganizationToken, error) {
	return a.db.getOrganizationTokenByID(ctx, tokenID)
}
