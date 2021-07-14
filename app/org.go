package app

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	db ots.OrganizationRepository
}

func NewOrganizationService(db ots.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		db: db,
	}
}

func (s OrganizationService) Create(opts *tfe.OrganizationCreateOptions) (*ots.Organization, error) {
	org, err := ots.NewOrganization(opts)
	if err != nil {
		return nil, err
	}

	return s.db.Create(org)
}

func (s OrganizationService) Get(name string) (*ots.Organization, error) {
	return s.db.Get(name)
}

func (s OrganizationService) List(opts tfe.OrganizationListOptions) (*ots.OrganizationList, error) {
	return s.db.List(opts)
}

func (s OrganizationService) Update(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
	return s.db.Update(name, func(org *ots.Organization) (err error) {
		org, err = ots.UpdateOrganization(org, opts)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s OrganizationService) Delete(name string) error {
	return s.db.Delete(name)
}

func (s OrganizationService) GetEntitlements(name string) (*ots.Entitlements, error) {
	org, err := s.db.Get(name)
	if err != nil {
		return nil, err
	}

	return ots.DefaultEntitlements(org.ExternalID), nil
}
