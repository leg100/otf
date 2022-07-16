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

func (s OrganizationService) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
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

func (s OrganizationService) EnsureCreated(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
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

func (s OrganizationService) List(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
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
