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
	org, err := otf.NewOrganization(opts)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Create(org)
	if err != nil {
		s.Error(err, "creating organization", "id", org.ID)
		return nil, err
	}

	s.es.Publish(otf.Event{Type: otf.OrganizationCreated, Payload: org})

	s.V(3).Info("created organization", "id", org.ID)

	return org, nil
}

func (s OrganizationService) Get(name string) (*otf.Organization, error) {
	org, err := s.db.Get(name)
	if err != nil {
		s.Error(err, "retrieving organization", "name", name)
		return nil, err
	}

	s.V(3).Info("retrieved organization", "name", name)

	return org, nil
}

func (s OrganizationService) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return s.db.List(opts)
}

func (s OrganizationService) Update(name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	return s.db.Update(name, func(org *otf.Organization) error {
		return otf.UpdateOrganization(org, opts)
	})
}

func (s OrganizationService) Delete(name string) error {
	return s.db.Delete(name)
}

func (s OrganizationService) GetEntitlements(name string) (*otf.Entitlements, error) {
	org, err := s.db.Get(name)
	if err != nil {
		return nil, err
	}

	return otf.DefaultEntitlements(org.ID), nil
}
