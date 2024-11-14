package team

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

var ErrRemovingOwnersTeamNotPermitted = errors.New("the owners team cannot be deleted")

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		db     *pgdb
		web    *webHandlers
		tfeapi *tfe
		api    *api

		afterCreateHooks []func(context.Context, *Team) error

		*teamTokenFactory
	}

	Options struct {
		*sql.DB
		*tfeapi.Responder
		html.Renderer
		logr.Logger

		OrganizationService *organization.Service
		TokensService       *tokens.Service
		Authorizer          *authz.Authorizer
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db:         &pgdb{opts.DB, opts.Logger},
		teamTokenFactory: &teamTokenFactory{
			tokens: opts.TokensService,
		},
	}
	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		tokens:   opts.TokensService,
		teams:    &svc,
	}
	svc.tfeapi = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}

	// Whenever an organization is created, also create an owners team. (The
	// user package hooks into CreateTeam to add the creator as a member).
	opts.OrganizationService.AfterCreateOrganization(func(ctx context.Context, organization *organization.Organization) error {
		// only an owner can create a team but there are no owners until an
		// owners team is created, so in this particular instance authorization
		// is skipped.
		ctx = authz.AddSkipAuthz(ctx)
		_, err := svc.Create(ctx, organization.Name, CreateTeamOptions{
			Name: internal.String("owners"),
		})
		if err != nil {
			return fmt.Errorf("creating owners team: %w", err)
		}
		return nil
	})
	// Register with auth middleware the team token kind and a means of
	// retrieving team corresponding to token.
	opts.TokensService.RegisterKind(TeamTokenKind, func(ctx context.Context, tokenID resource.ID) (authz.Subject, error) {
		return svc.GetTeamByTokenID(ctx, tokenID)
	})
	opts.Authorizer.RegisterOrganizationResolver(resource.TeamKind, func(ctx context.Context, id resource.ID) (string, error) {
		team, err := svc.db.getTeamByID(ctx, id)
		if err != nil {
			return "", err
		}
		return team.Organization, nil
	})

	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.tfeapi.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *Service) Create(ctx context.Context, organization string, opts CreateTeamOptions) (*Team, error) {
	subject, err := a.Authorize(ctx, rbac.CreateTeamAction, &authz.AccessRequest{Organization: organization})
	if err != nil {
		return nil, err
	}

	team, err := newTeam(organization, opts)
	if err != nil {
		return nil, err
	}

	err = a.db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		if err := a.db.createTeam(ctx, team); err != nil {
			return err
		}
		for _, hook := range a.afterCreateHooks {
			if err := hook(ctx, team); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		a.Error(err, "creating team", "name", team.Name, "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", team.Name, "organization", organization, "subject", subject)

	return team, nil
}

func (a *Service) AfterCreateTeam(hook func(context.Context, *Team) error) {
	a.afterCreateHooks = append(a.afterCreateHooks, hook)
}

func (a *Service) Update(ctx context.Context, teamID resource.ID, opts UpdateTeamOptions) (*Team, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}
	subject, err := a.Authorize(ctx, rbac.UpdateTeamAction, &authz.AccessRequest{Organization: team.Organization})
	if err != nil {
		return nil, err
	}

	team, err = a.db.UpdateTeam(ctx, teamID, func(team *Team) error {
		return team.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating team", "name", team.Name, "organization", team.Organization, "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated team", "name", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}

// List lists teams in the organization.
func (a *Service) List(ctx context.Context, organization string) ([]*Team, error) {
	subject, err := a.Authorize(ctx, rbac.ListTeamsAction, &authz.AccessRequest{Organization: organization})
	if err != nil {
		return nil, err
	}

	teams, err := a.db.listTeams(ctx, organization)
	if err != nil {
		a.Error(err, "listing teams", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed teams", "organization", organization, "subject", subject)

	return teams, nil
}

func (a *Service) Get(ctx context.Context, organization, name string) (*Team, error) {
	subject, err := a.Authorize(ctx, rbac.GetTeamAction, &authz.AccessRequest{Organization: organization})
	if err != nil {
		return nil, err
	}

	team, err := a.db.getTeam(ctx, name, organization)
	if err != nil {
		a.Error(err, "retrieving team", "team", name, "organization", organization, "subject", subject)
		return nil, err
	}

	a.V(9).Info("retrieved team", "team", name, "organization", organization, "subject", subject)

	return team, nil
}

func (a *Service) GetByID(ctx context.Context, teamID resource.ID) (*Team, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	subject, err := a.Authorize(ctx, rbac.GetTeamAction, &authz.AccessRequest{Organization: team.Organization})
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}

func (a *Service) Delete(ctx context.Context, teamID resource.ID) error {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.Authorize(ctx, rbac.DeleteTeamAction, &authz.AccessRequest{Organization: team.Organization})
	if err != nil {
		return err
	}

	if team.Name == "owners" {
		return ErrRemovingOwnersTeamNotPermitted
	}

	err = a.db.deleteTeam(ctx, teamID)
	if err != nil {
		a.Error(err, "deleting team", "team", team.Name, "organization", team.Organization, "subject", subject)
		return err
	}

	a.V(2).Info("deleted team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return nil
}

func (a *Service) GetTeamByTokenID(ctx context.Context, tokenID resource.ID) (*Team, error) {
	team, err := a.db.getTeamByTokenID(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving team by team token ID", "token_id", tokenID)
		return nil, err
	}

	subject, err := a.Authorize(ctx, rbac.GetTeamAction, &authz.AccessRequest{Organization: team.Organization})
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}
