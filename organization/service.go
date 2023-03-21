package organization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/rbac"
)

type (
	OrganizationService = Service

	Service interface {
		CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)
		UpdateOrganization(ctx context.Context, name string, opts OrganizationUpdateOptions) (*Organization, error)
		GetOrganization(ctx context.Context, name string) (*Organization, error)
		GetOrganizationJSONAPI(ctx context.Context, name string) (*jsonapi.Organization, error)
		ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error)
		DeleteOrganization(ctx context.Context, name string) error

		getEntitlements(ctx context.Context, organization string) (Entitlements, error)

		otf.Handlers
	}

	service struct {
		otf.Authorizer // authorize access to org
		logr.Logger
		pubsub.Broker

		api  *api
		db   *pgdb
		site otf.Authorizer // authorize access to site
		web  *web

		*jsonapiMarshaler
	}

	Options struct {
		otf.DB
		pubsub.Broker
		otf.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Authorizer:       &Authorizer{opts.Logger},
		Logger:           opts.Logger,
		Broker:           opts.Broker,
		db:               &pgdb{opts.DB},
		site:             &otf.SiteAuthorizer{opts.Logger},
		jsonapiMarshaler: &jsonapiMarshaler{},
	}
	svc.api = &api{
		svc:              &svc,
		jsonapiMarshaler: &jsonapiMarshaler{},
	}
	svc.web = &web{opts.Renderer, &svc}

	// Must register table name and service with pubsub broker so that it knows
	// how to lookup organizations in the DB.
	opts.Register("organization", &svc)

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *service) CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error) {
	subject, err := s.site.CanAccess(ctx, rbac.CreateOrganizationAction, "")
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	if err := s.db.create(ctx, org); err != nil {
		s.Error(err, "creating organization", "id", org.ID, "subject", subject)
		return nil, err
	}

	s.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	s.V(0).Info("created organization", "id", org.ID, "name", org.Name, "subject", subject)

	return org, nil
}

func (s *service) UpdateOrganization(ctx context.Context, name string, opts OrganizationUpdateOptions) (*Organization, error) {
	subject, err := s.CanAccess(ctx, rbac.UpdateOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := s.db.update(ctx, name, func(org *Organization) error {
		return org.Update(opts)
	})
	if err != nil {
		s.Error(err, "updating organization", "name", name, "subject", subject)
		return nil, err
	}

	s.V(2).Info("updated organization", "name", name, "id", org.ID, "subject", subject)

	return org, nil
}

// ListOrganizations lists organizations according to the subject. If the
// subject has site-wide permission to list organizations then all organizations
// are listed. Otherwise:
// Subject is a user: list their organization memberships
// Subject is an agent: return its organization
// Subject is an organization token: return its organization
// Subject is an team token: return its organization
func (s *service) ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error) {
	subject, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subject.CanAccessSite(rbac.ListOrganizationsAction) {
		return s.db.list(ctx, opts)
	}
	opts.Names = subject.ListOrganizations()
	return s.db.list(ctx, opts)
}

func (s *service) GetOrganization(ctx context.Context, name string) (*Organization, error) {
	subject, err := s.CanAccess(ctx, rbac.GetOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := s.db.get(ctx, name)
	if err != nil {
		s.Error(err, "retrieving organization", "name", name, "subject", subject)
		return nil, err
	}

	s.V(2).Info("retrieved organization", "name", name, "id", org.ID, "subject", subject)

	return org, nil
}

// GetByID implements pubsub.Getter
func (s *service) GetByID(ctx context.Context, id string) (any, error) {
	return s.db.getByID(ctx, id)
}

func (s *service) GetOrganizationJSONAPI(ctx context.Context, name string) (*jsonapi.Organization, error) {
	org, err := s.GetOrganization(ctx, name)
	if err != nil {
		return nil, err
	}
	return s.toOrganization(org), nil
}

func (s *service) DeleteOrganization(ctx context.Context, name string) error {
	subject, err := s.CanAccess(ctx, rbac.DeleteOrganizationAction, name)
	if err != nil {
		return err
	}

	err = s.db.delete(ctx, name)
	if err != nil {
		s.Error(err, "deleting organization", "name", name, "subject", subject)
		return err
	}
	s.V(0).Info("deleted organization", "name", name, "subject", subject)

	return nil
}

func (s *service) getEntitlements(ctx context.Context, organization string) (Entitlements, error) {
	_, err := s.CanAccess(ctx, rbac.GetEntitlementsAction, organization)
	if err != nil {
		return Entitlements{}, err
	}

	org, err := s.GetOrganization(ctx, organization)
	if err != nil {
		return Entitlements{}, err
	}
	return defaultEntitlements(org.ID), nil
}
