package team

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/hooks"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

var ErrRemovingOwnersTeamNotPermitted = errors.New("the owners team cannot be deleted")

type (
	TeamService interface {
		CreateTeam(ctx context.Context, organization string, opts CreateTeamOptions) (*Team, error)
		GetTeam(ctx context.Context, organization, team string) (*Team, error)
		GetTeamByID(ctx context.Context, teamID string) (*Team, error)
		GetTeamByTokenID(ctx context.Context, teamTokenID string) (*Team, error)
		ListTeams(ctx context.Context, organization string) ([]*Team, error)
		UpdateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error)
		DeleteTeam(ctx context.Context, teamID string) error

		AfterCreateTeam(l hooks.Listener[*Team])

		teamTokenService
	}

	service struct {
		logr.Logger

		organization internal.Authorizer // authorizes org access
		team         internal.Authorizer // authorizes team access

		db     *pgdb
		web    *webHandlers
		tfeapi *tfe
		api    *api

		createHook *hooks.Hook[*Team]

		*teamTokenFactory
	}

	Options struct {
		*sql.DB
		*tfeapi.Responder
		html.Renderer
		internal.HostnameService
		organization.OrganizationService
		tokens.TokensService
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		team:         &authorizer{Logger: opts.Logger},
		db:           &pgdb{opts.DB, opts.Logger},
		teamTokenFactory: &teamTokenFactory{
			TokensService: opts.TokensService,
		},
		createHook: hooks.NewHook[*Team](opts.DB),
	}
	svc.web = &webHandlers{
		Renderer:      opts.Renderer,
		tokensService: opts.TokensService,
		svc:           &svc,
	}
	svc.tfeapi = &tfe{
		TeamService: &svc,
		Responder:   opts.Responder,
	}
	svc.api = &api{
		TeamService: &svc,
		Responder:   opts.Responder,
	}

	// Whenever an organization is created, also create an owners team. (The
	// user package hooks into CreateTeam to add the creator as a member).
	opts.AfterCreateOrganization(func(ctx context.Context, organization *organization.Organization) error {
		// only an owner can create a team but there are no owners until an
		// owners team is created, so in this particular instance authorization
		// is skipped.
		ctx = internal.AddSkipAuthz(ctx)
		_, err := svc.CreateTeam(ctx, organization.Name, CreateTeamOptions{
			Name: internal.String("owners"),
		})
		if err != nil {
			return fmt.Errorf("creating owners team: %w", err)
		}
		return nil
	})
	// Register with auth middleware the team token kind and a means of
	// retrieving team corresponding to token.
	opts.RegisterKind(TeamTokenKind, func(ctx context.Context, tokenID string) (internal.Subject, error) {
		return svc.GetTeamByTokenID(ctx, tokenID)

	})

	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.tfeapi.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *service) AfterCreateTeam(l hooks.Listener[*Team]) {
	a.createHook.After(l)
}

func (a *service) CreateTeam(ctx context.Context, organization string, opts CreateTeamOptions) (*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateTeamAction, organization)
	if err != nil {
		return nil, err
	}

	team, err := newTeam(organization, opts)
	if err != nil {
		return nil, err
	}

	err = a.createHook.Dispatch(ctx, team, func(ctx context.Context) (*Team, error) {
		err := a.db.createTeam(ctx, team)
		return team, err
	})
	if err != nil {
		a.Error(err, "creating team", "name", team.Name, "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", team.Name, "organization", organization, "subject", subject)

	return team, nil
}

func (a *service) UpdateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}
	subject, err := a.organization.CanAccess(ctx, rbac.UpdateTeamAction, team.Organization)
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

// ListTeams lists teams in the organization.
func (a *service) ListTeams(ctx context.Context, organization string) ([]*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListTeamsAction, organization)
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

func (a *service) GetTeam(ctx context.Context, organization, name string) (*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, organization)
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

func (a *service) GetTeamByID(ctx context.Context, teamID string) (*Team, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, team.Organization)
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}

func (a *service) DeleteTeam(ctx context.Context, teamID string) error {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteTeamAction, team.Organization)
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

func (a *service) GetTeamByTokenID(ctx context.Context, tokenID string) (*Team, error) {
	team, err := a.db.getTeamByTokenID(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving team by team token ID", "token_id", tokenID)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, team.Organization)
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}
