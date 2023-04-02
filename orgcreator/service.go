package orgcreator

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type (
	OrganizationCreatorService = Service

	Service interface {
		CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*organization.Organization, error)
	}

	service struct {
		logr.Logger
		otf.Publisher

		api  *api
		db   otf.DB
		site otf.Authorizer // authorize access to site
		web  *web

		*organization.JSONAPIMarshaler
		auth.AuthService

		RestrictOrganizationCreation bool
	}

	Options struct {
		otf.DB
		otf.Publisher
		otf.Renderer
		logr.Logger

		auth.AuthService

		RestrictOrganizationCreation bool
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:                       opts.Logger,
		Publisher:                    opts.Publisher,
		JSONAPIMarshaler:             &organization.JSONAPIMarshaler{},
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		AuthService:                  opts.AuthService,
		db:                           opts.DB,
		site:                         &otf.SiteAuthorizer{Logger: opts.Logger},
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
	creator, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	err = s.db.Tx(ctx, func(tx otf.DB) error {
		_, err := tx.InsertOrganization(ctx, pggen.InsertOrganizationParams{
			ID:              sql.String(org.ID),
			CreatedAt:       sql.Timestamptz(org.CreatedAt),
			UpdatedAt:       sql.Timestamptz(org.UpdatedAt),
			Name:            sql.String(org.Name),
			SessionRemember: org.SessionRemember,
			SessionTimeout:  org.SessionTimeout,
		})
		if err != nil {
			return sql.Error(err)
		}

		// pre-emptively make the creator an owner to avoid a chicken-and-egg
		// problem when creating the owners team below: only an owner can create teams
		// but an owner can't be created until an owners team is created...
		creator.Teams = append(creator.Teams, &auth.Team{
			Name:         "owners",
			Organization: *opts.Name,
		})

		owners, err := s.AuthService.CreateTeam(ctx, auth.CreateTeamOptions{
			Name:         "owners",
			Organization: org.Name,
			Tx:           tx,
		})
		if err != nil {
			return fmt.Errorf("creating owners team: %w", err)
		}
		err = s.AuthService.AddTeamMembership(ctx, auth.TeamMembershipOptions{
			TeamID:   owners.ID,
			Username: creator.Username,
			Tx:       tx,
		})
		if err != nil {
			return fmt.Errorf("adding owner to owners team: %w", err)
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating organization", "id", org.ID, "subject", creator)
		return nil, err
	}

	s.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	s.V(0).Info("created organization", "id", org.ID, "name", org.Name, "subject", creator)

	return org, nil
}

func (s *service) authorize(ctx context.Context) (*auth.User, error) {
	subject, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := subject.(*auth.User)
	if !ok {
		s.Error(nil, "unauthorized action", "action", rbac.CreateOrganizationAction, "subject", subject)
		return nil, otf.ErrAccessNotPermitted
	}
	if s.RestrictOrganizationCreation && !user.IsSiteAdmin() {
		s.Error(nil, "unauthorized action", "action", rbac.CreateOrganizationAction, "subject", subject)
		return nil, otf.ErrAccessNotPermitted
	}
	return user, nil
}
