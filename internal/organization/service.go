package organization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

type (
	Service struct {
		RestrictOrganizationCreation bool

		*Authorizer // authorize access to org
		logr.Logger

		db           *pgdb
		site         authz.Authorizer // authorize access to site
		web          *web
		tfeapi       *tfe
		api          *api
		tokenFactory *tokenFactory
		broker       *pubsub.Broker[*Organization]

		afterCreateHooks  []func(context.Context, *Organization) error
		beforeDeleteHooks []func(context.Context, *Organization) error
	}

	Options struct {
		RestrictOrganizationCreation bool
		TokensService                *tokens.Service

		*sql.DB
		*tfeapi.Responder
		*sql.Listener
		html.Renderer
		logr.Logger
	}

	// ListOptions represents the options for listing organizations.
	ListOptions struct {
		resource.PageOptions
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Authorizer:                   &Authorizer{opts.Logger},
		Logger:                       opts.Logger,
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		db:                           &pgdb{opts.DB},
		site:                         &authz.SiteAuthorizer{Logger: opts.Logger},
		tokenFactory:                 &tokenFactory{tokens: opts.TokensService},
	}
	svc.web = &web{
		Renderer:         opts.Renderer,
		RestrictCreation: opts.RestrictOrganizationCreation,
		svc:              &svc,
	}
	svc.tfeapi = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}

	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.broker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"organizations",
		resource.OrganizationKind,
		func(ctx context.Context, id resource.ID, action sql.Action) (*Organization, error) {
			if action == sql.DeleteAction {
				return &Organization{ID: id}, nil
			}
			return svc.db.getByID(ctx, id)
		},
	)
	// Fetch organization when API calls request organization be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeOrganization, svc.tfeapi.include)
	// Register with auth middleware the organization token and a means of
	// retrieving organization corresponding to token.
	opts.TokensService.RegisterKind(resource.OrganizationTokenKind, func(ctx context.Context, tokenID resource.ID) (authz.Subject, error) {
		return svc.getOrganizationTokenByID(ctx, tokenID)
	})
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
}

func (s *Service) WatchOrganizations(ctx context.Context) (<-chan pubsub.Event[*Organization], func()) {
	return s.broker.Subscribe(ctx)
}

// Create creates an organization. Only users can create
// organizations, or, if RestrictOrganizationCreation is true, then only the
// site admin can create organizations. Creating an organization automatically
// creates an owners team and adds creator as an owner.
func (s *Service) Create(ctx context.Context, opts CreateOptions) (*Organization, error) {
	creator, err := s.restrictOrganizationCreation(ctx)
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	err = s.db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		if err := s.db.create(ctx, org); err != nil {
			return err
		}
		for _, hook := range s.afterCreateHooks {
			if err := hook(ctx, org); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating organization", "id", org.ID, "subject", creator)
		return nil, sql.Error(err)
	}
	s.V(0).Info("created organization", "id", org.ID, "name", org.Name, "subject", creator)

	return org, nil
}

func (s *Service) AfterCreateOrganization(hook func(context.Context, *Organization) error) {
	s.afterCreateHooks = append(s.afterCreateHooks, hook)
}

func (s *Service) Update(ctx context.Context, name string, opts UpdateOptions) (*Organization, error) {
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

// List lists organizations according to the subject. If the
// subject has site-wide permission to list organizations then all organizations
// are listed. Otherwise:
// Subject is a user: list their organization memberships
// Subject is an agent: return its organization
// Subject is an organization token: return its organization
// Subject is a team: return its organization
func (s *Service) List(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error) {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subject.CanAccessSite(rbac.ListOrganizationsAction) {
		return s.db.list(ctx, dbListOptions{PageOptions: opts.PageOptions})
	}
	return s.db.list(ctx, dbListOptions{PageOptions: opts.PageOptions, names: subject.Organizations()})
}

func (s *Service) Get(ctx context.Context, name string) (*Organization, error) {
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

func (s *Service) Delete(ctx context.Context, name string) error {
	subject, err := s.CanAccess(ctx, rbac.DeleteOrganizationAction, name)
	if err != nil {
		return err
	}

	err = s.db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		org, err := s.db.get(ctx, name)
		if err != nil {
			return err
		}
		for _, hook := range s.beforeDeleteHooks {
			if err := hook(ctx, org); err != nil {
				return err
			}
		}
		return s.db.delete(ctx, name)
	})
	if err != nil {
		s.Error(err, "deleting organization", "name", name, "subject", subject)
		return err
	}
	s.V(0).Info("deleted organization", "name", name, "subject", subject)

	return nil
}

func (s *Service) BeforeDeleteOrganization(hook func(context.Context, *Organization) error) {
	s.beforeDeleteHooks = append(s.beforeDeleteHooks, hook)
}

func (s *Service) GetEntitlements(ctx context.Context, organization string) (Entitlements, error) {
	_, err := s.CanAccess(ctx, rbac.GetEntitlementsAction, organization)
	if err != nil {
		return Entitlements{}, err
	}

	org, err := s.Get(ctx, organization)
	if err != nil {
		return Entitlements{}, err
	}
	return defaultEntitlements(org.ID), nil
}

func (s *Service) restrictOrganizationCreation(ctx context.Context) (authz.Subject, error) {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if s.RestrictOrganizationCreation && !subject.IsSiteAdmin() {
		s.Error(nil, "unauthorized action", "action", rbac.CreateOrganizationAction, "subject", subject)
		return subject, internal.ErrAccessNotPermitted
	}
	return subject, nil
}

// CreateToken creates an organization token. If an organization
// token already exists it is replaced.
func (s *Service) CreateToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	_, err := s.CanAccess(ctx, rbac.CreateOrganizationTokenAction, opts.Organization)
	if err != nil {
		return nil, nil, err
	}

	ot, token, err := s.tokenFactory.NewOrganizationToken(opts)
	if err != nil {
		s.Error(err, "constructing organization token", "organization", opts.Organization)
		return nil, nil, err
	}

	if err := s.db.upsertOrganizationToken(ctx, ot); err != nil {
		s.Error(err, "creating organization token", "organization", opts.Organization)
		return nil, nil, err
	}

	s.V(0).Info("created organization token", "organization", opts.Organization)

	return ot, token, nil
}

func (s *Service) GetOrganizationToken(ctx context.Context, organization string) (*OrganizationToken, error) {
	ot, err := s.db.getOrganizationTokenByName(ctx, organization)
	if err != nil {
		s.Error(err, "retrieving organization token", "organization", organization)
		return nil, err
	}
	s.V(0).Info("retrieved organization token", "organization", organization)
	return ot, nil
}

func (s *Service) getOrganizationTokenByID(ctx context.Context, tokenID resource.ID) (*OrganizationToken, error) {
	ot, err := s.db.getOrganizationTokenByID(ctx, tokenID)
	if err != nil {
		s.Error(err, "retrieving organization token", "token_id", tokenID)
		return nil, err
	}
	s.V(0).Info("retrieved organization token", "token_id", tokenID, "organization", ot.Organization)
	return ot, nil
}

func (s *Service) ListTokens(ctx context.Context, organization string) ([]*OrganizationToken, error) {
	tokens, err := s.db.listOrganizationTokens(ctx, organization)
	if err != nil {
		s.Error(err, "listing organization tokens", "organization", organization)
		return nil, err
	}
	s.V(0).Info("listed organization tokens", "organization", organization, "count", len(tokens))
	return tokens, nil
}

func (s *Service) DeleteToken(ctx context.Context, organization string) error {
	_, err := s.CanAccess(ctx, rbac.CreateOrganizationTokenAction, organization)
	if err != nil {
		return err
	}

	if err := s.db.deleteOrganizationToken(ctx, organization); err != nil {
		s.Error(err, "deleting organization token", "organization", organization)
		return err
	}

	s.V(0).Info("deleted organization token", "organization", organization)

	return nil
}
