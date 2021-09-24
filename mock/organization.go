package mock

import (
	"github.com/leg100/otf"
)

var _ otf.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	CreateOrganizationFn func(opts *otf.OrganizationCreateOptions) (*otf.Organization, error)
	UpdateOrganizationFn func(name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error)
	GetOrganizationFn    func(name string) (*otf.Organization, error)
	ListOrganizationFn   func(opts otf.OrganizationListOptions) (*otf.OrganizationList, error)
	DeleteOrganizationFn func(name string) error
	GetEntitlementsFn    func(name string) (*otf.Entitlements, error)
}

func (s OrganizationService) Create(opts *otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return s.CreateOrganizationFn(opts)
}

func (s OrganizationService) Get(name string) (*otf.Organization, error) {
	return s.GetOrganizationFn(name)
}

func (s OrganizationService) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return s.ListOrganizationFn(opts)
}

func (s OrganizationService) Update(name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	return s.UpdateOrganizationFn(name, opts)
}

func (s OrganizationService) Delete(name string) error {
	return s.DeleteOrganizationFn(name)
}

func (s OrganizationService) GetEntitlements(name string) (*otf.Entitlements, error) {
	return s.GetEntitlementsFn(name)
}

func NewOrganization(name, email string) *otf.Organization {
	return &otf.Organization{
		Name:                   name,
		Email:                  email,
		Permissions:            &otf.OrganizationPermissions{},
		SessionTimeout:         otf.DefaultSessionTimeout,
		SessionRemember:        otf.DefaultSessionExpiration,
		CollaboratorAuthPolicy: otf.DefaultCollaboratorAuthPolicy,
		CostEstimationEnabled:  otf.DefaultCostEstimationEnabled,
	}
}

func NewOrganizationList(name, email string, opts otf.OrganizationListOptions) *otf.OrganizationList {
	return &otf.OrganizationList{
		Items: []*otf.Organization{
			NewOrganization(name, email),
		},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}
}
