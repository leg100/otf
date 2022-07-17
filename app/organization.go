package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

var _ otf.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	db *sql.DB
	es otf.EventService

	logr.Logger
}

func NewOrganizationService(db *sql.DB, logger logr.Logger, es otf.EventService) (*OrganizationService, error) {
	svc := &OrganizationService{
		db:     db,
		es:     es,
		Logger: logger,
	}
	return svc, nil
}

// Create an organization. Needs admin permission.
func (s OrganizationService) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	if !otf.IsAdmin(ctx) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := otf.NewOrganization(opts)
	if err != nil {
		return nil, err
	}

	if err := s.db.CreateOrganization(ctx, org); err != nil {
		s.Error(err, "creating organization", "id", org.ID())
		return nil, err
	}

	s.es.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	s.V(0).Info("created organization", "id", org.ID(), "name", org.Name())

	return org, nil
}

// EnsureCreated idempotently creates an organization. Needs admin permission.
func (s OrganizationService) EnsureCreated(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	if !otf.IsAdmin(ctx) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := s.db.GetOrganization(ctx, *opts.Name)
	if err == nil {
		return org, nil
	}

	if err != otf.ErrResourceNotFound {
		s.Error(err, "retrieving organization", "name", *opts.Name)
		return nil, err
	}

	return s.Create(ctx, opts)
}

// Get retrieves an organization by name.
func (s OrganizationService) Get(ctx context.Context, name string) (*otf.Organization, error) {
	if !otf.CanAccess(ctx, &name) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := s.db.GetOrganization(ctx, name)
	if err != nil {
		s.Error(err, "retrieving organization", "name", name)
		return nil, err
	}

	s.V(2).Info("retrieved organization", "name", name, "id", org.ID())

	return org, nil
}

// List organizations. If the caller is a normal user then only list their
// organizations; otherwise list all.
func (s OrganizationService) List(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user, ok := subj.(*otf.User); ok && !user.SiteAdmin() {
		return newOrganizationList(opts, user.Organizations), nil
	}
	return s.db.ListOrganizations(ctx, opts)
}

func (s OrganizationService) Update(ctx context.Context, name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	if !otf.CanAccess(ctx, &name) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := s.db.UpdateOrganization(ctx, name, func(org *otf.Organization) error {
		return otf.UpdateOrganizationFromOpts(org, *opts)
	})
	if err != nil {
		s.Error(err, "updating organization", "name", name)
		return nil, err
	}

	s.V(2).Info("updated organization", "name", name, "id", org.ID())

	return org, nil
}

func (s OrganizationService) Delete(ctx context.Context, name string) error {
	if !otf.CanAccess(ctx, &name) {
		return otf.ErrAccessNotPermitted
	}

	err := s.db.DeleteOrganization(ctx, name)
	if err != nil {
		return err
	}
	return nil
}

func (s OrganizationService) GetEntitlements(ctx context.Context, organizationName string) (*otf.Entitlements, error) {
	if !otf.CanAccess(ctx, &organizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	org, err := s.Get(ctx, organizationName)
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
