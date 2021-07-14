package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	CreateOrganizationFn func(opts *tfe.OrganizationCreateOptions) (*ots.Organization, error)
	UpdateOrganizationFn func(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error)
	GetOrganizationFn    func(name string) (*ots.Organization, error)
	ListOrganizationFn   func(opts tfe.OrganizationListOptions) (*ots.OrganizationList, error)
	DeleteOrganizationFn func(name string) error
	GetEntitlementsFn    func(name string) (*ots.Entitlements, error)
}

func (s OrganizationService) Create(opts *tfe.OrganizationCreateOptions) (*ots.Organization, error) {
	return s.CreateOrganizationFn(opts)
}

func (s OrganizationService) Get(name string) (*ots.Organization, error) {
	return s.GetOrganizationFn(name)
}

func (s OrganizationService) List(opts tfe.OrganizationListOptions) (*ots.OrganizationList, error) {
	return s.ListOrganizationFn(opts)
}

func (s OrganizationService) Update(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
	return s.UpdateOrganizationFn(name, opts)
}

func (s OrganizationService) Delete(name string) error {
	return s.DeleteOrganizationFn(name)
}

func (s OrganizationService) GetEntitlements(name string) (*ots.Entitlements, error) {
	return s.GetEntitlementsFn(name)
}

func NewOrganization(name, email string) *ots.Organization {
	return &ots.Organization{
		Name:                   name,
		Email:                  email,
		Permissions:            &tfe.OrganizationPermissions{},
		SessionTimeout:         ots.DefaultSessionTimeout,
		SessionRemember:        ots.DefaultSessionExpiration,
		CollaboratorAuthPolicy: ots.DefaultCollaboratorAuthPolicy,
		CostEstimationEnabled:  ots.DefaultCostEstimationEnabled,
	}
}

func NewOrganizationList(name, email string, opts tfe.OrganizationListOptions) *ots.OrganizationList {
	return &ots.OrganizationList{
		Items: []*ots.Organization{
			NewOrganization(name, email),
		},
		Pagination: ots.NewPagination(opts.ListOptions, 1),
	}
}
