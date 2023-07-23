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
		ID           string
		CreatedAt    time.Time
		Organization string // Token belongs to an organization
	}

	// CreateOrganizationTokenOptions are options for creating an organization token via the service
	// endpoint
	CreateOrganizationTokenOptions struct {
		Organization string `schema:"organization_name,required"`
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
		// DeleteOrganizationToken deletes a user token.
		DeleteOrganizationToken(ctx context.Context, tokenID string) error
	}
)

func NewOrganizationToken(opts NewOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	ot := OrganizationToken{
		ID:           internal.NewID("ot"),
		CreatedAt:    internal.CurrentTimestamp(),
		Organization: opts.Organization,
	}
	token, err := NewToken(NewTokenOptions{
		key:     opts.key,
		Subject: ot.ID,
		Kind:    organizationTokenKind,
	})
	if err != nil {
		return nil, nil, err
	}
	return &ot, token, nil
}

func (u *OrganizationToken) CanAccessSite(action rbac.Action) bool {
	// an organization token can only be used for intra-organization resources
	return false
}

func (u *OrganizationToken) CanAccessOrganization(action rbac.Action, org string) bool {
	if u.Organization != org {
		return false
	}
	switch action {
	// permit team management
	case rbac.CreateTeamAction, rbac.UpdateTeamAction, rbac.GetTeamAction, rbac.ListTeamsAction, rbac.DeleteTeamAction, rbac.AddTeamMembershipAction, rbac.RemoveTeamMembershipAction:
		return true
	// permit workspace management
	case rbac.ListWorkspacesAction, rbac.GetWorkspaceAction, rbac.CreateWorkspaceAction, rbac.DeleteWorkspaceAction, rbac.SetWorkspacePermissionAction, rbac.UnsetWorkspacePermissionAction, rbac.UpdateWorkspaceAction:
		return true
	// permit tag management
	case rbac.ListTagsAction, rbac.DeleteTagsAction, rbac.TagWorkspacesAction, rbac.AddTagsAction, rbac.RemoveTagsAction, rbac.ListWorkspaceTags:
		return true
	default:
		return false
	}
}

func (u *OrganizationToken) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	return u.CanAccessOrganization(action, policy.Organization)
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

	a.V(1).Info("created organization token", "organization", opts.Organization)

	return ot, token, nil
}

func (a *service) GetOrganizationToken(ctx context.Context, organization string) (*OrganizationToken, error) {
	return a.db.getOrganizationToken(ctx, organization)
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

	a.V(1).Info("deleted organization token", "organization", organization)

	return nil
}
