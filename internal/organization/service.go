package organization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/hooks"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

type (
	OrganizationService = Service

	Service interface {
		CreateOrganization(ctx context.Context, opts CreateOptions) (*Organization, error)
		UpdateOrganization(ctx context.Context, name string, opts UpdateOptions) (*Organization, error)
		GetOrganization(ctx context.Context, name string) (*Organization, error)
		ListOrganizations(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error)
		DeleteOrganization(ctx context.Context, name string) error
		GetEntitlements(ctx context.Context, organization string) (Entitlements, error)
		AfterCreateOrganization(l hooks.Listener[*Organization])
		BeforeDeleteOrganization(l hooks.Listener[*Organization])

		// organization tokens
		CreateOrganizationToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error)
		// GetOrganizationToken gets the organization token. If a token does not
		// exist, then nil is returned without an error.
		GetOrganizationToken(ctx context.Context, organization string) (*OrganizationToken, error)
		DeleteOrganizationToken(ctx context.Context, organization string) error
		WatchOrganizations(context.Context) (<-chan pubsub.Event[*Organization], func())
		getOrganizationTokenByID(ctx context.Context, tokenID string) (*OrganizationToken, error)
	}

	service struct {
		RestrictOrganizationCreation bool

		internal.Authorizer // authorize access to org
		logr.Logger

		db           *pgdb
		site         internal.Authorizer // authorize access to site
		web          *web
		tfeapi       *tfe
		api          *api
		tokenFactory *tokenFactory
		broker       *pubsub.Broker[*Organization]

		createHook *hooks.Hook[*Organization]
		deleteHook *hooks.Hook[*Organization]
	}

	Options struct {
		RestrictOrganizationCreation bool

		*sql.DB
		*tfeapi.Responder
		*sql.Listener
		html.Renderer
		logr.Logger
		tokens.TokensService
	}

	// ListOptions represents the options for listing organizations.
	ListOptions struct {
		resource.PageOptions
	}
)

func NewService(opts Options) *service {
	svc := service{
		Authorizer:                   &Authorizer{opts.Logger},
		Logger:                       opts.Logger,
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		db:                           &pgdb{opts.DB},
		site:                         &internal.SiteAuthorizer{Logger: opts.Logger},
		createHook:                   hooks.NewHook[*Organization](opts.DB),
		deleteHook:                   hooks.NewHook[*Organization](opts.DB),
		tokenFactory:                 &tokenFactory{TokensService: opts.TokensService},
	}
	svc.web = &web{
		Renderer:                     opts.Renderer,
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		svc:                          &svc,
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
		func(ctx context.Context, id string, action sql.Action) (*Organization, error) {
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
	opts.TokensService.RegisterKind(OrganizationTokenKind, func(ctx context.Context, organization string) (internal.Subject, error) {
		return svc.GetOrganizationToken(ctx, organization)
	})

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.web.addHandlers(r)
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
}

func (s *service) AfterCreateOrganization(l hooks.Listener[*Organization]) {
	s.createHook.After(l)
}

func (s *service) BeforeDeleteOrganization(l hooks.Listener[*Organization]) {
	s.deleteHook.Before(l)
}

func (s *service) WatchOrganizations(ctx context.Context) (<-chan pubsub.Event[*Organization], func()) {
	return s.broker.Subscribe(ctx)
}

// CreateOrganization creates an organization. Only users can create
// organizations, or, if RestrictOrganizationCreation is true, then only the
// site admin can create organizations. Creating an organization automatically
// creates an owners team and adds creator as an owner.
func (s *service) CreateOrganization(ctx context.Context, opts CreateOptions) (*Organization, error) {
	creator, err := s.restrictOrganizationCreation(ctx)
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	err = s.createHook.Dispatch(ctx, org, func(ctx context.Context) (*Organization, error) {
		_, err = s.db.Conn(ctx).InsertOrganization(ctx, pggen.InsertOrganizationParams{
			ID:                         sql.String(org.ID),
			CreatedAt:                  sql.Timestamptz(org.CreatedAt),
			UpdatedAt:                  sql.Timestamptz(org.UpdatedAt),
			Name:                       sql.String(org.Name),
			SessionRemember:            sql.Int4Ptr(org.SessionRemember),
			SessionTimeout:             sql.Int4Ptr(org.SessionTimeout),
			Email:                      sql.StringPtr(org.Email),
			CollaboratorAuthPolicy:     sql.StringPtr(org.CollaboratorAuthPolicy),
			CostEstimationEnabled:      sql.Bool(org.CostEstimationEnabled),
			AllowForceDeleteWorkspaces: sql.Bool(org.AllowForceDeleteWorkspaces),
		})
		return org, sql.Error(err)
	})
	if err != nil {
		s.Error(err, "creating organization", "id", org.ID, "subject", creator)
		return nil, sql.Error(err)
	}
	s.V(0).Info("created organization", "id", org.ID, "name", org.Name, "subject", creator)

	return org, nil
}

func (s *service) UpdateOrganization(ctx context.Context, name string, opts UpdateOptions) (*Organization, error) {
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
// Subject is a team: return its organization
func (s *service) ListOrganizations(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subject.CanAccessSite(rbac.ListOrganizationsAction) {
		return s.db.list(ctx, dbListOptions{PageOptions: opts.PageOptions})
	}
	return s.db.list(ctx, dbListOptions{PageOptions: opts.PageOptions, names: subject.Organizations()})
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

	org := &Organization{Name: name}
	err = s.deleteHook.Dispatch(ctx, org, func(ctx context.Context) (*Organization, error) {
		return org, s.db.delete(ctx, name)
	})
	if err != nil {
		s.Error(err, "deleting organization", "name", name, "subject", subject)
		return err
	}
	s.V(0).Info("deleted organization", "name", name, "subject", subject)

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

func (s *service) restrictOrganizationCreation(ctx context.Context) (internal.Subject, error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if s.RestrictOrganizationCreation && !subject.IsSiteAdmin() {
		s.Error(nil, "unauthorized action", "action", rbac.CreateOrganizationAction, "subject", subject)
		return subject, internal.ErrAccessNotPermitted
	}
	return subject, nil
}

// CreateOrganizationToken creates an organization token. If an organization
// token already exists it is replaced.
func (s *service) CreateOrganizationToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
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

func (s *service) GetOrganizationToken(ctx context.Context, organization string) (*OrganizationToken, error) {
	return s.db.getOrganizationTokenByName(ctx, organization)
}

func (s *service) DeleteOrganizationToken(ctx context.Context, organization string) error {
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

func (s *service) getOrganizationTokenByID(ctx context.Context, tokenID string) (*OrganizationToken, error) {
	return s.db.getOrganizationTokenByID(ctx, tokenID)
}
