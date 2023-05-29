package organization

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
)

type (
	OrganizationService = Service

	Service interface {
		UpdateOrganization(ctx context.Context, name string, opts OrganizationUpdateOptions) (*Organization, error)
		GetOrganization(ctx context.Context, name string) (*Organization, error)
		ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error)
		DeleteOrganization(ctx context.Context, name string) error

		GetEntitlements(ctx context.Context, organization string) (Entitlements, error)
	}

	service struct {
		internal.Authorizer // authorize access to org
		logr.Logger
		*pubsub.Broker

		db                           *pgdb
		site                         internal.Authorizer // authorize access to site
		web                          *web
		RestrictOrganizationCreation bool
	}

	Options struct {
		internal.DB
		*pubsub.Broker
		html.Renderer
		logr.Logger

		RestrictOrganizationCreation bool
	}
)

func NewService(opts Options) *service {
	svc := service{
		Authorizer:                   &Authorizer{opts.Logger},
		Logger:                       opts.Logger,
		Broker:                       opts.Broker,
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		db:                           &pgdb{opts.DB},
		site:                         &internal.SiteAuthorizer{Logger: opts.Logger},
	}
	svc.web = &web{
		Renderer:                     opts.Renderer,
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		svc:                          &svc,
	}

	// Register with broker an unmarshaler for unmarshaling organization
	// database table events into organization events.
	opts.Register("organizations", svc.db)

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
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

	s.Publish(pubsub.NewUpdatedEvent(org))
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
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subject.CanAccessSite(rbac.ListOrganizationsAction) {
		return s.db.list(ctx, opts)
	}
	opts.Names = subject.Organizations()
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

	s.V(9).Info("retrieved organization", "name", name, "subject", subject)

	return org, nil
}

func (s *service) DeleteOrganization(ctx context.Context, name string) error {
	subject, err := s.CanAccess(ctx, rbac.DeleteOrganizationAction, name)
	if err != nil {
		return err
	}

	org, err := s.db.get(ctx, name)
	if err != nil {
		s.Error(err, "retrieving organization", "name", name, "subject", subject)
		return err
	}

	err = s.db.delete(ctx, name)
	if err != nil {
		s.Error(err, "deleting organization", "name", name, "subject", subject)
		return err
	}
	s.V(0).Info("deleted organization", "name", name, "subject", subject)

	s.Publish(pubsub.NewDeletedEvent(org))

	return nil
}

func (s *service) GetEntitlements(ctx context.Context, organization string) (Entitlements, error) {
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
