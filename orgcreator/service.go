package orgcreator

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
)

type (
	OrganizationCreatorService = Service

	Service interface {
		CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*organization.Organization, error)
	}

	service struct {
		logr.Logger
		otf.Broker

		api  *api
		db   *pgdb
		site otf.Authorizer // authorize access to site
		web  *web

		*organization.JSONAPIMarshaler

		RestrictOrganizationCreation bool
	}

	Options struct {
		otf.DB
		otf.Broker
		otf.Renderer
		logr.Logger

		RestrictOrganizationCreation bool
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:                       opts.Logger,
		Broker:                       opts.Broker,
		db:                           &pgdb{opts.DB},
		site:                         &otf.SiteAuthorizer{opts.Logger},
		JSONAPIMarshaler:             &organization.JSONAPIMarshaler{},
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
	}
	svc.api = &api{
		svc:              &svc,
		JSONAPIMarshaler: &organization.JSONAPIMarshaler{},
	}
	svc.web = &web{opts.Renderer, &svc}

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// CreateOrganization creates an organization. Only users can create
// organizations, or, if RestrictOrganizationCreation is true, then only the
// site admin can create organizations. Creating an organization automatically
// creates an owners team and adds creator as an owner.
func (s *service) CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*organization.Organization, error) {
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
