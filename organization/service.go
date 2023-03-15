package organization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

type (
	Service interface {
		CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
		GetOrganization(ctx context.Context, name string) (*Organization, error)
		GetOrganizationJSONAPI(ctx context.Context, name string) (*jsonapi.Organization, error)

		create(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
		get(ctx context.Context, name string) (*Organization, error)
		list(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error)
		update(ctx context.Context, name string, opts OrganizationUpdateOptions) (*Organization, error)
		delete(ctx context.Context, name string) error
		getEntitlements(ctx context.Context, organization string) (Entitlements, error)
	}

	service struct {
		otf.Authorizer // authorize access to org
		logr.Logger
		otf.Publisher

		api  *api
		db   *pgdb
		site otf.Authorizer // authorize access to site
		web  *web
	}

	Options struct {
		otf.DB
		otf.Publisher
		otf.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Authorizer: &Authorizer{opts.Logger},
		Logger:     opts.Logger,
		Publisher:  opts.Publisher,
	}
	svc.api = &api{&svc}
	svc.db = newDB(opts.DB)
	svc.site = &otf.SiteAuthorizer{opts.Logger}
	svc.web = &web{opts.Renderer, &svc}
	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (a *service) CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error) {
	return a.create(ctx, opts)
}

func (a *service) UpdateOrganization(ctx context.Context, name string, opts OrganizationUpdateOptions) (*Organization, error) {
	return a.update(ctx, name, opts)
}

func (a *service) ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error) {
	return a.list(ctx, opts)
}

func (a *service) GetOrganization(ctx context.Context, name string) (*Organization, error) {
	return a.get(ctx, name)
}

func (a *service) GetOrganizationJSONAPI(ctx context.Context, name string) (*jsonapi.Organization, error) {
	return nil, nil
}

func (a *service) DeleteOrganization(ctx context.Context, name string) error {
	return a.delete(ctx, name)
}

func (a *service) create(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error) {
	subject, err := a.site.CanAccess(ctx, rbac.CreateOrganizationAction, "")
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	if err := a.db.create(ctx, org); err != nil {
		a.Error(err, "creating organization", "id", org.ID, "subject", subject)
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	a.V(0).Info("created organization", "id", org.ID, "name", org.Name, "subject", subject)

	return org, nil
}

func (a *service) get(ctx context.Context, name string) (*Organization, error) {
	subject, err := a.CanAccess(ctx, rbac.GetOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := a.db.get(ctx, name)
	if err != nil {
		a.Error(err, "retrieving organization", "name", name, "subject", subject)
		return nil, err
	}

	a.V(2).Info("retrieved organization", "name", name, "id", org.ID, "subject", subject)

	return org, nil
}

// list lists organizations across site.
func (a *service) list(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error) {
	_, err := a.site.CanAccess(ctx, rbac.ListOrganizationsAction, "")
	if err != nil {
		return nil, err
	}
	return a.db.list(ctx, opts)
}

func (a *service) listByUser(ctx context.Context, userID string, opts OrganizationListOptions) (*OrganizationList, error) {
	if userID == otf.SiteAdminID {
		// site admin gets all orgs across site.
		return a.db.list(ctx, opts)
	}
	return a.db.listByUser(ctx, userID, opts)
}

func (a *service) update(ctx context.Context, name string, opts OrganizationUpdateOptions) (*Organization, error) {
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

	a.V(2).Info("updated organization", "name", name, "id", org.ID, "subject", subject)

	return org, nil
}

func (a *service) delete(ctx context.Context, name string) error {
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

func (a *service) getEntitlements(ctx context.Context, organization string) (Entitlements, error) {
	_, err := a.CanAccess(ctx, rbac.GetEntitlementsAction, organization)
	if err != nil {
		return Entitlements{}, err
	}

	org, err := a.get(ctx, organization)
	if err != nil {
		return Entitlements{}, err
	}
	return defaultEntitlements(org.ID), nil
}
