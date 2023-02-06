package organization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// appService is the application service for organizations
type appService interface {
	CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (string, error)
	EnsureCreatedOrganization(ctx context.Context, opts OrganizationCreateOptions) (string, error)
	GetOrganization(ctx context.Context, name string) (*Organization, error)
	ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*organizationList, error)
	UpdateOrganization(ctx context.Context, name string, opts *OrganizationUpdateOptions) (*Organization, error)
	DeleteOrganization(ctx context.Context, name string) error

	createOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
}

// app is the implementation of appService
type app struct {
	otf.Authorizer
	logr.Logger
	otf.PubSubService

	db
}

// CreateOrganization creates an organization. Needs admin permission.
func (a *app) CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (string, error) {
	org, err := a.createOrganization(ctx, opts)
	if err != nil {
		return "", err
	}
	return org.Name(), nil
}

// EnsureCreatedOrganization idempotently creates an organization. Needs admin
// permission.
//
// TODO: merge this into CreatedOrganization and add an option to toggle
// idempotency
func (a *app) EnsureCreatedOrganization(ctx context.Context, opts OrganizationCreateOptions) (string, error) {
	subject, err := a.CanAccessSite(ctx, rbac.GetOrganizationAction)
	if err != nil {
		return "", err
	}

	_, err = a.db.GetOrganization(ctx, *opts.Name)
	if err == nil {
		return "", nil
	}

	if err != otf.ErrResourceNotFound {
		a.Error(err, "retrieving organization", "name", *opts.Name, "subject", subject)
		return "", err
	}

	// org not found

	return a.CreateOrganization(ctx, opts)
}

// createOrganization creates an organization. Needs admin permission.
func (a *app) createOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error) {
	subject, err := a.CanAccessSite(ctx, rbac.CreateOrganizationAction)
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	if err := a.db.CreateOrganization(ctx, org); err != nil {
		a.Error(err, "creating organization", "id", org.ID(), "subject", subject)
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	a.V(0).Info("created organization", "id", org.ID(), "name", org.Name(), "subject", subject)

	return org, nil
}

// GetOrganization retrieves an organization by name.
func (a *app) GetOrganization(ctx context.Context, name string) (*Organization, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.GetOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := a.db.GetOrganization(ctx, name)
	if err != nil {
		a.Error(err, "retrieving organization", "name", name, "subject", subject)
		return nil, err
	}

	a.V(2).Info("retrieved organization", "name", name, "id", org.ID(), "subject", subject)

	return org, nil
}

// ListOrganizations lists organizations. If the caller is a normal user then
// only list their organizations; otherwise list all.
func (a *app) ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*organizationList, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user, ok := subj.(*otf.User); ok && !user.IsSiteAdmin() {
		orgs, err := a.db.ListOrganizationsByUser(ctx, user.ID())
		if err != nil {
			return nil, err
		}
		return newOrganizationList(opts, orgs), nil
	}
	return a.db.ListOrganizations(ctx, opts)
}

func (a *app) UpdateOrganization(ctx context.Context, name string, opts *OrganizationUpdateOptions) (*Organization, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.UpdateOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := a.db.UpdateOrganization(ctx, name, func(org *Organization) error {
		return org.Update(*opts)
	})
	if err != nil {
		a.Error(err, "updating organization", "name", name, "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated organization", "name", name, "id", org.ID(), "subject", subject)

	return org, nil
}

func (a *app) DeleteOrganization(ctx context.Context, name string) error {
	subject, err := a.CanAccessOrganization(ctx, rbac.DeleteOrganizationAction, name)
	if err != nil {
		return err
	}

	err = a.db.DeleteOrganization(ctx, name)
	if err != nil {
		a.Error(err, "deleting organization", "name", name, "subject", subject)
		return err
	}
	a.V(0).Info("deleted organization", "name", name, "subject", subject)

	return nil
}

func (a *app) GetEntitlements(ctx context.Context, organization string) (*otf.Entitlements, error) {
	_, err := a.CanAccessOrganization(ctx, rbac.GetEntitlementsAction, organization)
	if err != nil {
		return nil, err
	}

	org, err := a.GetOrganization(ctx, organization)
	if err != nil {
		return nil, err
	}
	return otf.DefaultEntitlements(org.ID()), nil
}

// newOrganizationList constructs a paginated OrganizationList given the list
// options and a complete list of organizations.
func newOrganizationList(opts OrganizationListOptions, orgs []*Organization) *organizationList {
	low := opts.GetOffset()
	if low > len(orgs) {
		low = len(orgs)
	}
	high := opts.GetOffset() + opts.GetLimit()
	if high > len(orgs) {
		high = len(orgs)
	}
	return &organizationList{
		Items:      orgs[low:high],
		Pagination: otf.NewPagination(opts.ListOptions, len(orgs)),
	}
}
