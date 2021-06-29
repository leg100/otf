package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	CreateOrganizationFn func(opts *tfe.OrganizationCreateOptions) (*tfe.Organization, error)
	UpdateOrganizationFn func(name string, opts *tfe.OrganizationUpdateOptions) (*tfe.Organization, error)
	GetOrganizationFn    func(name string) (*tfe.Organization, error)
	ListOrganizationFn   func(opts ots.OrganizationListOptions) (*ots.OrganizationList, error)
	DeleteOrganizationFn func(name string) error
	GetEntitlementsFn    func(name string) (*tfe.Entitlements, error)
}

func (s OrganizationService) CreateOrganization(opts *tfe.OrganizationCreateOptions) (*tfe.Organization, error) {
	return s.CreateOrganizationFn(opts)
}

func (s OrganizationService) UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*tfe.Organization, error) {
	return s.UpdateOrganizationFn(name, opts)
}

func (s OrganizationService) GetOrganization(name string) (*tfe.Organization, error) {
	return s.GetOrganizationFn(name)
}

func (s OrganizationService) ListOrganizations(opts ots.OrganizationListOptions) (*ots.OrganizationList, error) {
	return s.ListOrganizationFn(opts)
}

func (s OrganizationService) DeleteOrganization(name string) error {
	return s.DeleteOrganizationFn(name)
}

func (s OrganizationService) GetEntitlements(name string) (*tfe.Entitlements, error) {
	return s.GetEntitlementsFn(name)
}

func NewOrganization(name, email string) *tfe.Organization {
	return &tfe.Organization{
		Name:                   name,
		Email:                  email,
		Permissions:            &tfe.OrganizationPermissions{},
		SessionTimeout:         ots.DefaultSessionTimeout,
		SessionRemember:        ots.DefaultSessionExpiration,
		CollaboratorAuthPolicy: ots.DefaultCollaboratorAuthPolicy,
		CostEstimationEnabled:  ots.DefaultCostEstimationEnabled,
	}
}

func NewOrganizationList(name, email string, opts ots.OrganizationListOptions) *ots.OrganizationList {
	return &ots.OrganizationList{
		Items: []*tfe.Organization{
			NewOrganization(name, email),
		},
		Pagination: ots.NewPagination(opts.ListOptions, 1),
	}
}
