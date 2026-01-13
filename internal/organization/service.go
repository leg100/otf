package organization

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

type (
	Service struct {
		RestrictOrganizationCreation bool

		*authz.Authorizer
		logr.Logger

		db           *pgdb
		tfeapi       *tfe
		api          *api
		tokenFactory *tokenFactory

		afterCreateHooks  []func(context.Context, *Organization) error
		beforeDeleteHooks []func(context.Context, *Organization) error
	}

	Options struct {
		RestrictOrganizationCreation bool
		TokensService                *tokens.Service
		Authorizer                   *authz.Authorizer

		*sql.DB
		*tfeapi.Responder
		*sql.Listener
		logr.Logger
	}

	// ListOptions represents the options for listing organizations.
	ListOptions struct {
		resource.PageOptions
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Authorizer:                   opts.Authorizer,
		Logger:                       opts.Logger,
		RestrictOrganizationCreation: opts.RestrictOrganizationCreation,
		db:                           &pgdb{opts.DB},
		tokenFactory:                 &tokenFactory{tokens: opts.TokensService},
	}
	svc.tfeapi = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}

	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}
	// Fetch organization when API calls request organization be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeOrganization, svc.tfeapi.include)
	// Register with auth middleware the organization token and a means of
	// retrieving organization corresponding to token.
	opts.TokensService.RegisterKind(resource.OrganizationTokenKind, func(ctx context.Context, tokenID resource.TfeID) (authz.Subject, error) {
		return svc.getOrganizationTokenByID(ctx, tokenID)
	})
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
}

// Create creates an organization. Only users can create
// organizations, or, if RestrictOrganizationCreation is true, then only the
// site admin can create organizations. Creating an organization automatically
// creates an owners team and adds creator as an owner.
func (s *Service) Create(ctx context.Context, opts CreateOptions) (*Organization, error) {
	subject, err := s.Authorize(ctx, authz.CreateOrganizationAction, resource.SiteID)
	if err != nil {
		return nil, err
	}
	if err := s.restrictOrganizationCreation(subject); err != nil {
		return nil, err
	}
	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}
	err = s.db.Tx(ctx, func(ctx context.Context) error {
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
		s.Error(err, "creating organization", "id", org.ID, "subject", subject)
		return nil, err
	}
	s.V(0).Info("created organization", "id", org.ID, "name", org.Name, "subject", subject)
	return org, nil
}

func (s *Service) restrictOrganizationCreation(subject authz.Subject) error {
	if s.RestrictOrganizationCreation {
		type user interface {
			IsSiteAdmin() bool
		}
		if user, ok := subject.(user); !ok || !user.IsSiteAdmin() {
			s.Error(internal.ErrAccessNotPermitted, "cannot create organization because creation is restricted to site admins", "action", authz.CreateOrganizationAction, "subject", subject)
			return internal.ErrAccessNotPermitted
		}
	}
	return nil
}

func (s *Service) AfterCreateOrganization(hook func(context.Context, *Organization) error) {
	s.afterCreateHooks = append(s.afterCreateHooks, hook)
}

func (s *Service) Update(ctx context.Context, name Name, opts UpdateOptions) (*Organization, error) {
	subject, err := s.Authorize(ctx, authz.UpdateOrganizationAction, &name)
	if err != nil {
		return nil, err
	}
	org, err := s.db.update(ctx, name, func(ctx context.Context, org *Organization) error {
		return org.Update(opts)
	})
	if err != nil {
		s.Error(err, "updating organization", "name", name, "subject", subject)
		return nil, err
	}

	s.V(2).Info("updated organization", "name", name, "id", org.ID, "subject", subject)
	return org, nil
}

// List organizations. If the subject lacks the ListOrganizationsAction
// permission then its organization memberships are listed instead.
func (s *Service) List(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error) {
	orgs, subject, err := func() (*resource.Page[*Organization], authz.Subject, error) {
		var names []Name
		subject, err := s.Authorize(ctx, authz.ListOrganizationsAction, resource.SiteID, authz.WithoutErrorLogging())
		if errors.Is(err, internal.ErrAccessNotPermitted) {
			// List subject's organization memberships instead.
			type memberships interface {
				Organizations() []Name
			}
			user, ok := subject.(memberships)
			if !ok {
				return nil, subject, err
			}
			names = user.Organizations()
		} else if err != nil {
			return nil, subject, err
		}
		orgs, err := s.db.list(ctx, dbListOptions{PageOptions: opts.PageOptions, names: names})
		return orgs, subject, err
	}()
	if err != nil {
		s.Error(err, "listing organizations", "subject", subject)
	}
	s.V(9).Info("listed organizations", "subject", subject)
	return orgs, err
}

func (s *Service) Get(ctx context.Context, name Name) (*Organization, error) {
	subject, err := s.Authorize(ctx, authz.GetOrganizationAction, &name)
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

func (s *Service) Delete(ctx context.Context, name Name) error {
	subject, err := s.Authorize(ctx, authz.DeleteOrganizationAction, &name)
	if err != nil {
		return err
	}

	err = s.db.Tx(ctx, func(ctx context.Context) error {
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

func (s *Service) GetEntitlements(ctx context.Context, organization Name) (Entitlements, error) {
	_, err := s.Authorize(ctx, authz.GetEntitlementsAction, organization)
	if err != nil {
		return Entitlements{}, err
	}

	org, err := s.Get(ctx, organization)
	if err != nil {
		return Entitlements{}, err
	}
	return defaultEntitlements(org.ID), nil
}

// CreateToken creates an organization token. If an organization
// token already exists it is replaced.
func (s *Service) CreateToken(ctx context.Context, opts CreateOrganizationTokenOptions) (*OrganizationToken, []byte, error) {
	_, err := s.Authorize(ctx, authz.CreateOrganizationTokenAction, &opts.Organization)
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

func (s *Service) GetOrganizationToken(ctx context.Context, organization Name) (*OrganizationToken, error) {
	ot, err := s.db.getOrganizationTokenByName(ctx, organization)
	if err != nil {
		s.Error(err, "retrieving organization token", "organization", organization)
		return nil, err
	}
	s.V(0).Info("retrieved organization token", "organization", organization)
	return ot, nil
}

func (s *Service) getOrganizationTokenByID(ctx context.Context, tokenID resource.TfeID) (*OrganizationToken, error) {
	ot, err := s.db.getOrganizationTokenByID(ctx, tokenID)
	if err != nil {
		s.Error(err, "retrieving organization token", "token_id", tokenID)
		return nil, err
	}
	s.V(0).Info("retrieved organization token", "token_id", tokenID, "organization", ot.Organization)
	return ot, nil
}

func (s *Service) ListTokens(ctx context.Context, organization Name) ([]*OrganizationToken, error) {
	tokens, err := s.db.listOrganizationTokens(ctx, organization)
	if err != nil {
		s.Error(err, "listing organization tokens", "organization", organization)
		return nil, err
	}
	s.V(0).Info("listed organization tokens", "organization", organization, "count", len(tokens))
	return tokens, nil
}

func (s *Service) DeleteToken(ctx context.Context, organization Name) error {
	_, err := s.Authorize(ctx, authz.CreateOrganizationTokenAction, organization)
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
