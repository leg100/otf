package auth

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/rbac"
)

type app interface {
	// CreateRegistrySession creates a registry session, returning its token.
	CreateRegistrySession(ctx context.Context, organization string) (string, error)

	createAgentToken(ctx context.Context, options otf.CreateAgentTokenOptions) (*agentToken, error)
	listAgentTokens(ctx context.Context, organization string) ([]*agentToken, error)
	deleteAgentToken(ctx context.Context, id string) (*agentToken, error)

	listUsers(context.Context, UserListOptions) ([]*User, error)

	getTeam(ctx context.Context, teamID string) (*Team, error)
	listTeams(ctx context.Context, organization string) ([]*Team, error)
	listTeamMembers(ctx context.Context, teamID string) ([]*User, error)
	updateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error)

	createSession(ctx context.Context, organization string) (*Session, error)
	getSession(ctx context.Context, token string) (*Session, error)
	listSessions(ctx context.Context, userID string) ([]*Session, error)
	deleteSession(ctx context.Context, token string) error
}

type Application struct {
	otf.Authorizer
	logr.Logger

	db db
	*handlers
	*synchroniser
}

func NewApplication(ctx context.Context, opts ApplicationOptions) (*Application, error) {
	db := newDB(opts.Database)
	app := &Application{
		Authorizer:   opts.Authorizer,
		Logger:       opts.Logger,
		db:           db,
		synchroniser: &synchroniser{Logger, db},
	}

	authenticators, err := newAuthenticators(opts.Logger, opts.Application, opts.Configs)
	if err != nil {
		return nil, err
	}

	app.handlers = &handlers{
		app:            app,
		authenticators: authenticators,
	}

	// purge expired registry sessions
	go db.startExpirer(ctx, defaultExpiry)

	return app
}

type ApplicationOptions struct {
	Configs []*cloud.CloudOAuthConfig

	otf.Authorizer
	otf.Database
	logr.Logger
}

func NewApp(logger logr.Logger, db otf.DB, authorizer otf.Authorizer) *Application {
	return &Application{
		Logger:     logger,
		db:         newDB(db, logger),
		Authorizer: authorizer,
	}
}

func (a *Application) CreateUser(ctx context.Context, username string) (otf.User, error) {
	user := NewUser(username)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username)
		return nil, err
	}

	a.V(0).Info("created user", "username", username)

	return user, nil
}

func (a *Application) GetUser(ctx context.Context, spec UserSpec) (otf.User, error) {
	user, err := a.db.GetUser(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving user", "spec", spec)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username())

	return user, nil
}

func (a *Application) CreateTeam(ctx context.Context, opts CreateTeamOptions) (*Team, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateTeamAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	team := newTeam(opts)

	if err := a.db.CreateTeam(ctx, team); err != nil {
		a.Error(err, "creating team", "name", opts.Name, "organization", opts.Organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", opts.Name, "organization", opts.Organization, "subject", subject)

	return team, nil
}

// listUsers lists an organization's users
func (a *Application) listUsers(ctx context.Context, organization string) ([]*User, error) {
	_, err := a.CanAccessOrganization(ctx, rbac.ListUsersAction, organization)
	if err != nil {
		return nil, err
	}

	return a.db.listUsers(ctx, organization)
}

//
// Session endpoints
//

func (a *Application) createSession(r *http.Request, userID string) (*Session, error) {
	session, err := NewSession(r, userID)
	if err != nil {
		a.Error(err, "building new session", "uid", userID)
		return nil, err
	}
	if err := a.db.CreateSession(r.Context(), session); err != nil {
		a.Error(err, "creating session", "uid", userID)
		return nil, err
	}

	a.V(2).Info("created session", "uid", userID)

	return session, nil
}

func (a *Application) getSession(ctx context.Context, token string) (*Session, error) {
	return a.db.GetSessionByToken(ctx, token)
}

func (a *Application) listSessions(ctx context.Context, userID string) ([]*Session, error) {
	return a.db.ListSessions(ctx, userID)
}

func (a *Application) deleteSession(ctx context.Context, token string) error {
	if err := a.db.DeleteSession(ctx, token); err != nil {
		a.Error(err, "deleting session")
		return err
	}

	a.V(2).Info("deleted session")

	return nil
}

func (a *Application) updateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}
	subject, err := a.CanAccessOrganization(ctx, rbac.UpdateTeamAction, team.Organization())
	if err != nil {
		return nil, err
	}

	team, err = a.db.UpdateTeam(ctx, teamID, func(team *Team) error {
		return team.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating team", "name", team.Name(), "organization", team.Organization(), "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated team", "name", team.Name(), "organization", team.Organization(), "subject", subject)

	return team, nil
}

// GetTeam retrieves a team.
func (a *Application) getTeam(ctx context.Context, teamID string) (*Team, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	// Check organization-wide authority
	subject, err := a.CanAccessOrganization(ctx, rbac.GetTeamAction, team.Organization())
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved team", "team", team.Name(), "organization", team.Organization(), "subject", subject)

	return team, nil
}

// listTeams lists teams in the organization.
func (a *Application) listTeams(ctx context.Context, organization string) ([]*Team, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.ListTeamsAction, organization)
	if err != nil {
		return nil, err
	}

	teams, err := a.db.ListTeams(ctx, organization)
	if err != nil {
		a.V(2).Info("listing teams", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed teams", "organization", organization, "subject", subject)

	return teams, nil
}

// listTeamMembers lists users that are members of the given team. The caller
// needs either organization-wide authority to call this endpoint, or they need
// to be a member of the team.
func (a *Application) listTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, rbac.ListUsersAction, team.Organization())
	if err != nil {
		return nil, err
	}

	members, err := a.db.listTeamMembers(ctx, teamID)
	if err != nil {
		a.Error(err, "listing team members", "team_id", teamID, "subject", subject)
		return nil, err
	}

	a.V(2).Info("listed team members", "team_id", teamID, "subject", subject)

	return members, nil
}

// Registry session services

// CreateRegistrySession creates and persists a registry session.
func (a *Application) CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error) {
	return a.createRegistrySession(ctx, organization)
}

// GetRegistrySession retrieves a registry session using a token. Useful for
// checking token is valid.
func (a *Application) getRegistrySession(ctx context.Context, token string) (*registrySession, error) {
	// No need for authz because caller is providing an auth token.

	session, err := a.db.getRegistrySession(ctx, token)
	if err != nil {
		a.Error(err, "retrieving registry session", "token", "*****")
		return nil, err
	}

	a.V(2).Info("retrieved registry session", "session", session)

	return session, nil
}

func (a *Application) createRegistrySession(ctx context.Context, organization string) (*registrySession, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateRegistrySessionAction, organization)
	if err != nil {
		return nil, err
	}

	session, err := newRegistrySession(organization)
	if err != nil {
		a.Error(err, "constructing registry session", "subject", subject, "organization", organization)
		return nil, err
	}
	if err := a.db.createRegistrySession(ctx, session); err != nil {
		a.Error(err, "creating registry session", "subject", subject, "session", session)
		return nil, err
	}

	a.V(2).Info("created registry session", "subject", subject, "session", session)

	return session, nil
}

// AgentToken services

func (a *Application) createAgentToken(ctx context.Context, opts CreateAgentTokenOptions) (*agentToken, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateAgentTokenAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	token, err := newAgentToken(opts)
	if err != nil {
		return nil, err
	}
	if err := a.db.CreateAgentToken(ctx, token); err != nil {
		a.Error(err, "creating agent token", "organization", opts.Organization, "id", token.ID(), "subject", subject)
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.Organization, "id", token.ID(), "subject", subject)
	return token, nil
}

func (a *Application) listAgentTokens(ctx context.Context, organization string) ([]*agentToken, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.ListAgentTokensAction, organization)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.ListAgentTokens(ctx, organization)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed agent tokens", "organization", organization, "subject", subject)
	return tokens, nil
}

// GetAgentToken retrieves an agent token using the given token.
func (a *Application) GetAgentToken(ctx context.Context, token string) (*otf.AgentToken, error) {
	at, err := a.db.GetAgentTokenByToken(ctx, token)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}
	a.V(2).Info("retrieved agent token", "organization", at.Organization(), "id", at.ID())
	return at, nil
}

func (a *Application) deleteAgentToken(ctx context.Context, id string) (*agentToken, error) {
	// retrieve agent token first in order to get organization for authorization
	at, err := a.db.GetAgentTokenByID(ctx, id)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}

	subject, err := a.CanAccessOrganization(ctx, rbac.DeleteAgentTokenAction, at.Organization())
	if err != nil {
		return nil, err
	}

	if err := a.db.deleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "agent token", at, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted agent token", "agent token", at, "subject", subject)
	return at, nil
}
