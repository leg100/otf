package app

import (
	"context"

	"github.com/leg100/otf"
)

// CreateOrganization creates an organization. Needs admin permission.
func (a *Application) CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	if !otf.IsAdmin(ctx) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := otf.NewOrganization(opts)
	if err != nil {
		return nil, err
	}

	if err := a.db.CreateOrganization(ctx, org); err != nil {
		a.Error(err, "creating organization", "id", org.ID())
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	a.V(0).Info("created organization", "id", org.ID(), "name", org.Name())

	return org, nil
}

// EnsureCreatedOrganization idempotently creates an organization. Needs admin
// permission.
func (a *Application) EnsureCreatedOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	if !otf.IsAdmin(ctx) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := a.db.GetOrganization(ctx, *opts.Name)
	if err == nil {
		return org, nil
	}

	if err != otf.ErrResourceNotFound {
		a.Error(err, "retrieving organization", "name", *opts.Name)
		return nil, err
	}

	return a.CreateOrganization(ctx, opts)
}

// GetOrganization retrieves an organization by name.
func (a *Application) GetOrganization(ctx context.Context, name string) (*otf.Organization, error) {
	if !otf.CanAccess(ctx, &name) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := a.db.GetOrganization(ctx, name)
	if err != nil {
		a.Error(err, "retrieving organization", "name", name)
		return nil, err
	}

	a.V(2).Info("retrieved organization", "name", name, "id", org.ID())

	return org, nil
}

// ListOrganization organizations. If the caller is a normal user then only list
// their organizations; otherwise list all.
func (a *Application) ListOrganization(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user, ok := subj.(*otf.User); ok && !user.SiteAdmin() {
		return newOrganizationList(opts, user.Organizations), nil
	}
	return a.db.ListOrganizations(ctx, opts)
}

func (a *Application) UpdateOrganization(ctx context.Context, name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	if !otf.CanAccess(ctx, &name) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := a.db.UpdateOrganization(ctx, name, func(org *otf.Organization) error {
		return otf.UpdateOrganizationFromOpts(org, *opts)
	})
	if err != nil {
		a.Error(err, "updating organization", "name", name)
		return nil, err
	}

	a.V(2).Info("updated organization", "name", name, "id", org.ID())

	return org, nil
}

func (a *Application) DeleteOrganization(ctx context.Context, name string) error {
	if !otf.CanAccess(ctx, &name) {
		return otf.ErrAccessNotPermitted
	}

	err := a.db.DeleteOrganization(ctx, name)
	if err != nil {
		return err
	}
	return nil
}

func (a *Application) GetEntitlements(ctx context.Context, organizationName string) (*otf.Entitlements, error) {
	if !otf.CanAccess(ctx, &organizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := a.GetOrganization(ctx, organizationName)
	if err != nil {
		return nil, err
	}
	return otf.DefaultEntitlements(org.ID()), nil
}

// newOrganizationList constructs a paginated OrganizationList given the list
// options and a complete list of organizations.
func newOrganizationList(opts otf.OrganizationListOptions, orgs []*otf.Organization) *otf.OrganizationList {
	low := opts.GetOffset()
	if low > len(orgs) {
		low = len(orgs)
	}
	high := opts.GetOffset() + opts.GetLimit()
	if high > len(orgs) {
		high = len(orgs)
	}
	return &otf.OrganizationList{
		Items:      orgs[low:high],
		Pagination: otf.NewPagination(opts.ListOptions, len(orgs)),
	}
}
