package organization

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type Service struct {
	application
	api *api
	web *web
}

func NewService(opts Options) *Service {
	app := &app{
		Authorizer: opts.Authorizer,
		Logger:     opts.Logger,
		db:         newDB(opts.DB),
	}
	return &Service{
		application: app,
		api:         &api{app},
	}
}

type Options struct {
	otf.Authorizer
	otf.DB
	otf.Renderer
	logr.Logger
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (a *Service) CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (otf.Organization, error) {
	org, err := a.create(ctx, opts)
	if err != nil {
		return otf.Organization{}, nil
	}
	return org.toValue(), nil
}

func (a *Service) UpdateOrganization(ctx context.Context, name string, opts UpdateOptions) (otf.Organization, error) {
	org, err := a.update(ctx, name, opts)
	if err != nil {
		return otf.Organization{}, nil
	}
	return org.toValue(), nil
}

func (a *Service) ListOrganizations(ctx context.Context, opts ListOptions) (otf.OrganizationList, error) {
	from, err := a.list(ctx, opts)
	if err != nil {
		return otf.OrganizationList{}, nil
	}
	to := otf.OrganizationList{
		Pagination: from.Pagination,
	}
	for _, org := range from.Items {
		to.Items = append(to.Items, org.toValue())
	}
	return to, nil
}

func (a *Service) GetOrganization(ctx context.Context, name string) (otf.Organization, error) {
	org, err := a.get(ctx, name)
	if err != nil {
		return otf.Organization{}, nil
	}
	return org.toValue(), nil
}

func (a *Service) DeleteOrganization(ctx context.Context, name string) error {
	return a.delete(ctx, name)
}
