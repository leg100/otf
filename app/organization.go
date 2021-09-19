package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	db otf.OrganizationStore
	es otf.EventService
}

func NewOrganizationService(db otf.OrganizationStore, es otf.EventService) *OrganizationService {
	return &OrganizationService{
		db: db,
		es: es,
	}
}

func (s OrganizationService) Create(opts *tfe.OrganizationCreateOptions) (*otf.Organization, error) {
	org, err := otf.NewOrganization(opts)
	if err != nil {
		return nil, err
	}

	org, err = s.db.Create(org)
	if err != nil {
		return nil, err
	}

	s.es.Publish(otf.Event{Type: otf.OrganizationCreated, Payload: org})

	return org, nil
}

func (s OrganizationService) Get(name string) (*otf.Organization, error) {
	return s.db.Get(name)
}

func (s OrganizationService) List(opts tfe.OrganizationListOptions) (*otf.OrganizationList, error) {
	return s.db.List(opts)
}

func (s OrganizationService) Update(name string, opts *tfe.OrganizationUpdateOptions) (*otf.Organization, error) {
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
