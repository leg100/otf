package organization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type service interface {
	create(ctx context.Context, opts otf.OrganizationCreateOptions) (*Organization, error)
	get(ctx context.Context, name string) (*Organization, error)
	list(ctx context.Context, opts ListOptions) (*OrganizationList, error)
	update(ctx context.Context, name string, opts UpdateOptions) (*Organization, error)
	delete(ctx context.Context, name string) error
	getEntitlements(ctx context.Context, organization string) (Entitlements, error)
}

type Service struct {
	*Authorizer // authorize access to org
	logr.Logger
	otf.PubSubService

	api  *api
	db   *pgdb
	site otf.Authorizer // authorize access to site
	web  *web
}

func NewService(opts Options) *Service {
	svc := Service{
		Authorizer:    &Authorizer{opts.Logger},
		Logger:        opts.Logger,
		PubSubService: opts.PubSubService,
	}
	svc.api = &api{&svc}
	svc.db = newDB(opts.DB)
	svc.site = &otf.SiteAuthorizer{opts.Logger}
	svc.web = &web{opts.Renderer, &svc}
	return &svc
}

type Options struct {
	otf.DB
	otf.PubSubService
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

func (a *Service) create(ctx context.Context, opts otf.OrganizationCreateOptions) (*Organization, error) {
	subject, err := a.site.CanAccess(ctx, rbac.CreateOrganizationAction, "")
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	if err := a.db.create(ctx, org); err != nil {
		a.Error(err, "creating organization", "id", org.ID(), "subject", subject)
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	a.V(0).Info("created organization", "id", org.ID(), "name", org.Name(), "subject", subject)

	return org, nil
}

func (a *Service) get(ctx context.Context, name string) (*Organization, error) {
	subject, err := a.CanAccess(ctx, rbac.GetOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := a.db.get(ctx, name)
	if err != nil {
		a.Error(err, "retrieving organization", "name", name, "subject", subject)
		return nil, err
	}

	a.V(2).Info("retrieved organization", "name", name, "id", org.ID(), "subject", subject)

	return org, nil
}

func (a *Service) list(ctx context.Context, opts ListOptions) (*OrganizationList, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user, ok := subj.(otf.User); ok && !user.IsSiteAdmin() {
		return a.db.listByUser(ctx, user.ID(), opts)
	}
	return a.db.list(ctx, opts)
}

func (a *Service) update(ctx context.Context, name string, opts UpdateOptions) (*Organization, error) {
	subject, err := a.CanAccess(ctx, rbac.UpdateOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := a.db.update(ctx, name, func(org *Organization) error {
		return org.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating organization", "name", name, "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated organization", "name", name, "id", org.ID(), "subject", subject)

	return org, nil
}

func (a *Service) delete(ctx context.Context, name string) error {
	subject, err := a.CanAccess(ctx, rbac.DeleteOrganizationAction, name)
	if err != nil {
		return err
	}

	err = a.db.delete(ctx, name)
	if err != nil {
		a.Error(err, "deleting organization", "name", name, "subject", subject)
		return err
	}
	a.V(0).Info("deleted organization", "name", name, "subject", subject)

	return nil
}

func (a *Service) getEntitlements(ctx context.Context, organization string) (Entitlements, error) {
	_, err := a.CanAccess(ctx, rbac.GetEntitlementsAction, organization)
	if err != nil {
		return Entitlements{}, err
	}

	org, err := a.get(ctx, organization)
	if err != nil {
		return Entitlements{}, err
	}
	return defaultEntitlements(org.ID()), nil
}
