package orgcreator

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	OrganizationCreatorService = Service

	Service interface {
		CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*organization.Organization, error)
	}

	service struct {
		logr.Logger
		pubsub.Publisher

		db   *sql.DB
		site internal.Authorizer // authorize access to site
		web  *web

		auth.AuthService

		RestrictOrganizationCreation bool
	}

	Options struct {
		*sql.DB
		pubsub.Publisher
		html.Renderer
		logr.Logger

		auth.AuthService

		RestrictOrganizationCreation bool
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:                       opts.Logger,
		Publisher:                    opts.Publisher,
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		AuthService:                  opts.AuthService,
		db:                           opts.DB,
		site:                         &internal.SiteAuthorizer{Logger: opts.Logger},
	}
	svc.web = &web{opts.Renderer, &svc}

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
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

	err = s.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertOrganization(ctx, pggen.InsertOrganizationParams{
			ID:                     sql.String(org.ID),
			CreatedAt:              sql.Timestamptz(org.CreatedAt),
			UpdatedAt:              sql.Timestamptz(org.UpdatedAt),
			Name:                   sql.String(org.Name),
			SessionRemember:        sql.Int4Ptr(org.SessionRemember),
			SessionTimeout:         sql.Int4Ptr(org.SessionTimeout),
			Email:                  sql.StringPtr(org.Email),
			CollaboratorAuthPolicy: sql.StringPtr(org.CollaboratorAuthPolicy),
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

		owners, err := s.AuthService.CreateTeam(ctx, org.Name, auth.CreateTeamOptions{
			Name: internal.String("owners"),
		})
		if err != nil {
			return fmt.Errorf("creating owners team: %w", err)
		}
		err = s.AuthService.AddTeamMembership(ctx, auth.TeamMembershipOptions{
			TeamID:    owners.ID,
			Usernames: []string{creator.Username},
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

	s.V(0).Info("created organization", "id", org.ID, "name", org.Name, "subject", creator)

	return org, nil
}

func (s *service) authorize(ctx context.Context) (*auth.User, error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := subject.(*auth.User)
	if !ok {
		s.Error(nil, "unauthorized action", "action", rbac.CreateOrganizationAction, "subject", subject)
		return nil, internal.ErrAccessNotPermitted
	}
	if s.RestrictOrganizationCreation && !user.IsSiteAdmin() {
		s.Error(nil, "unauthorized action", "action", rbac.CreateOrganizationAction, "subject", subject)
		return nil, internal.ErrAccessNotPermitted
	}
	return user, nil
}
