package mock

import (
	"context"

	"github.com/leg100/otf"
)

var _ otf.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	CreateOrganizationFn func(opts otf.OrganizationCreateOptions) (*otf.Organization, error)
	UpdateOrganizationFn func(name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error)
	GetOrganizationFn    func(name string) (*otf.Organization, error)
	ListOrganizationFn   func(opts otf.OrganizationListOptions) (*otf.OrganizationList, error)
	DeleteOrganizationFn func(name string) error
	GetEntitlementsFn    func(name string) (*otf.Entitlements, error)
}

func (s OrganizationService) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return s.CreateOrganizationFn(opts)
}

func (s OrganizationService) Get(ctx context.Context, name string) (*otf.Organization, error) {
	return s.GetOrganizationFn(name)
}

func (s OrganizationService) List(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return s.ListOrganizationFn(opts)
}

func (s OrganizationService) Update(ctx context.Context, name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	return s.UpdateOrganizationFn(name, opts)
}

func (s OrganizationService) Delete(ctx context.Context, name string) error {
	return s.DeleteOrganizationFn(name)
}

func (s OrganizationService) GetEntitlements(ctx context.Context, name string) (*otf.Entitlements, error) {
	return s.GetEntitlementsFn(name)
}

func NewOrganization(name string) *otf.Organization {
	return &otf.Organization{
		Name:            name,
		SessionTimeout:  otf.DefaultSessionTimeout,
		SessionRemember: otf.DefaultSessionExpiration,
	}
}

func NewOrganizationList(name string, opts otf.OrganizationListOptions) *otf.OrganizationList {
	return &otf.OrganizationList{
		Items: []*otf.Organization{
			NewOrganization(name),
		},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}
}
