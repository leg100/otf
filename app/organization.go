package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	db otf.OrganizationStore
	es otf.EventService

	logr.Logger
}

func NewOrganizationService(db otf.OrganizationStore, logger logr.Logger, es otf.EventService) *OrganizationService {
	return &OrganizationService{
		db:     db,
		es:     es,
		Logger: logger,
	}
}

func (s OrganizationService) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	// TODO: check whether org already exists first

	org, err := otf.NewOrganization(opts)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Create(org)
	if err != nil {
		s.Error(err, "creating organization", "id", org.ID)
		return nil, err
	}

	s.es.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	s.V(0).Info("created organization", "id", org.ID, "name", org.Name)

	return org, nil
}

func (s OrganizationService) EnsureCreated(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	org, err := s.db.Get(*opts.Name)
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
	org, err := s.db.Get(name)
	if err != nil {
		s.Error(err, "retrieving organization", "name", name)
		return nil, err
	}

	s.V(2).Info("retrieved organization", "name", name, "id", org.ID)

	return org, nil
}

func (s OrganizationService) List(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return s.db.List(opts)
}

func (s OrganizationService) Update(ctx context.Context, name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	return s.db.Update(name, func(org *otf.Organization) error {
		return otf.UpdateOrganizationFromOpts(org, *opts)
	})
}

func (s OrganizationService) Delete(ctx context.Context, name string) error {
	return s.db.Delete(name)
}

func (s OrganizationService) GetEntitlements(ctx context.Context, name string) (*otf.Entitlements, error) {
	org, err := s.db.Get(name)
	if err != nil {
		return nil, err
	}

	return otf.DefaultEntitlements(org.ID), nil
}
