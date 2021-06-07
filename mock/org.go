package mock

import (
	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	CreateOrganizationFn func(name string, opts *tfe.OrganizationCreateOptions) (*ots.Organization, error)
	UpdateOrganizationFn func(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error)
	GetOrganizationFn    func(name string) (*ots.Organization, error)
	ListOrganizationFn   func() ([]*ots.Organization, error)
	DeleteOrganizationFn func(name string) error
	GetEntitlementsFn    func(name string) (*ots.Entitlements, error)
}

func (s OrganizationService) CreateOrganization(name string, opts *tfe.OrganizationCreateOptions) (*ots.Organization, error) {
	return s.CreateOrganizationFn(name, opts)
}

func (s OrganizationService) UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
	return s.UpdateOrganizationFn(name, opts)
}

func (s OrganizationService) GetOrganization(name string) (*ots.Organization, error) {
	return s.GetOrganizationFn(name)
}

func (s OrganizationService) ListOrganizations() ([]*ots.Organization, error) {
	return s.ListOrganizationFn()
}

func (s OrganizationService) DeleteOrganization(name string) error {
	return s.DeleteOrganizationFn(name)
}

func (s OrganizationService) GetEntitlements(name string) (*ots.Entitlements, error) {
	return s.GetEntitlementsFn(name)
}
