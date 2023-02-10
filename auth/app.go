package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type Application struct {
	otf.Authorizer
	logr.Logger

	db db
}

type service interface {
	create(ctx context.Context, organization string) (*Session, error)
	get(ctx context.Context, token string) (*Session, error)
}

// app is the implementation of service
type app struct {
	otf.Authorizer
	logr.Logger

	db
	*handlers
}

func NewApplication(ctx context.Context, opts ApplicationOptions) *app {
	app := &app{
		Authorizer: opts.Authorizer,
		Logger:     opts.Logger,
		db:         newDB(ctx, opts.Database, defaultExpiry),
	}
	app.handlers = &handlers{
		app: app,
	}
	return app
}

type ApplicationOptions struct {
	otf.Authorizer
	otf.Database
	logr.Logger
}

func NewApp(logger logr.Logger, db otf.DB, authorizer otf.Authorizer) *Application {
	return &Application{
		Logger:     logger,
		db:         newPGDB(db),
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

// EnsureCreatedUser retrieves the user or creates the user if they don't exist.
func (a *Application) EnsureCreatedUser(ctx context.Context, username string) (otf.User, error) {
	user, err := a.db.GetUser(ctx, UserSpec{Username: &username})
	if err == nil {
		return user, nil
	}
	if err != otf.ErrResourceNotFound {
		a.Error(err, "retrieving user", "username", username)
		return nil, err
	}

	return a.CreateUser(ctx, username)
}

// ListUsers lists users with various filters.
func (a *Application) ListUsers(ctx context.Context, opts UserListOptions) ([]otf.User, error) {
	var err error
	if opts.Organization != nil && opts.TeamName != nil {
		// team members can view membership of their own teams.
		subject, err := otf.SubjectFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if user, ok := subject.(otf.User); ok && !user.IsSiteAdmin() {
			_, err := user.Team(*opts.TeamName, *opts.Organization)
			if err != nil {
				// user is not a member of the team
				return nil, otf.ErrAccessNotPermitted
			}
			return a.db.ListUsers(ctx, opts)
		}
	}

	if opts.Organization != nil {
		// subject needs perms on org to list users in org
		_, err = a.CanAccessOrganization(ctx, rbac.ListUsersAction, *opts.Organization)
	} else {
		// subject needs perms on site to list users across site
		_, err = a.CanAccessSite(ctx, rbac.ListRunsAction)
	}
	if err != nil {
		return nil, err
	}

	return a.db.ListUsers(ctx, opts)
}

func (a *Application) SyncUserMemberships(ctx context.Context, user *otf.User, orgs []string, teams []*otf.Team) (*otf.User, error) {
	if err := user.SyncMemberships(ctx, a.db, orgs, teams); err != nil {
		return nil, err
	}

	a.V(1).Info("synchronised user's memberships", "username", user.Username())

	return user, nil
}

func (a *Application) GetUser(ctx context.Context, spec UserSpec) (otf.User, error) {
	user, err := a.db.GetUser(ctx, spec)
	if err != nil {
		// Failure to retrieve a user is frequently due to the fact the http
		// middleware first calls this endpoint to see if the bearer token
		// belongs to a user and if that fails it then checks if it belongs to
		// an agent. Therefore we log this at a low priority info level rather
		// than as an error.
		//
		// TODO: make bearer token a signed cryptographic token containing
		// metadata about the authenticating entity.
		info := append([]any{"err", err.Error()}, spec.KeyValue()...)
		a.V(2).Info("unable to retrieve user", info...)
		return nil, err
	}

	a.V(2).Info("retrieved user", "username", user.Username())

	return user, nil
}

func (a *Application) CreateTeam(ctx context.Context, opts otf.CreateTeamOptions) (*otf.Team, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateTeamAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	org, err := a.db.GetOrganization(ctx, opts.Organization)
	if err != nil {
		return nil, err
	}

	team := newTeam(opts.Name, org)

	if err := a.db.CreateTeam(ctx, team); err != nil {
		a.Error(err, "creating team", "name", opts.Name, "organization", opts.Organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", opts.Name, "organization", opts.Organization, "subject", subject)

	return team, nil
}

// EnsureCreatedTeam retrieves the team or creates the team if it doesn't exist.
func (a *Application) EnsureCreatedTeam(ctx context.Context, opts otf.CreateTeamOptions) (*otf.Team, error) {
	team, err := a.db.GetTeam(ctx, opts.Name, opts.Organization)
	if err == otf.ErrResourceNotFound {
		return a.CreateTeam(ctx, opts)
	} else if err != nil {
		a.Error(err, "retrieving team", "name", opts.Name, "organization", opts.Organization)
		return nil, err
	} else {
		return team, nil
	}
}

func (a *Application) UpdateTeam(ctx context.Context, teamID string, opts otf.UpdateTeamOptions) (*otf.Team, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}
	subject, err := a.CanAccessOrganization(ctx, rbac.UpdateTeamAction, team.Organization())
	if err != nil {
		return nil, err
	}

	team, err = a.db.UpdateTeam(ctx, teamID, func(team *otf.Team) error {
		return team.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating team", "name", team.Name(), "organization", team.Organization(), "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated team", "name", team.Name(), "organization", team.Organization(), "subject", subject)

	return team, nil
}

// GetTeam retrieves a team in an organization. If the caller is an unprivileged
// user i.e. not an owner nor a site admin then they are only permitted to
// retrieve a team they are a member of.
func (a *Application) GetTeam(ctx context.Context, teamID string) (*otf.Team, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	// Check organization-wide authority
	subject, err := a.CanAccessOrganization(ctx, rbac.GetTeamAction, team.Organization())
	if err != nil {
		// Fallback to checking if they are member of the team
		if user, ok := subject.(*otf.User); ok {
			if !user.IsTeamMember(teamID) {
				// User is not a member; refuse access
				return nil, err
			}
		} else {
			// non-user without organization-wide authority; refuse access
			return nil, err
		}
	}

	a.V(2).Info("retrieved team", "team", team.Name(), "organization", team.Organization(), "subject", subject)

	return team, nil
}

// ListTeams lists teams in the organization. If the caller is an unprivileged
// user i.e. not an owner nor a site admin then only their teams are listed.
func (a *Application) ListTeams(ctx context.Context, organization string) ([]*otf.Team, error) {
	if user, err := otf.UserFromContext(ctx); err == nil && user.IsUnprivilegedUser(organization) {
		a.V(2).Info("listed teams", "organization", organization, "subject", user)
		return user.TeamsByOrganization(organization), nil
	}

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

// ListTeamMembers lists users that are members of the given team. The caller
// needs either organization-wide authority to call this endpoint, or they need
// to be a member of the team.
func (a *Application) ListTeamMembers(ctx context.Context, teamID string) ([]*otf.User, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	// Check organization-wide authority
	subject, err := a.CanAccessOrganization(ctx, rbac.ListTeamsAction, team.Organization())
	if err != nil {
		// Fallback to checking if they are member of the team
		if user, ok := subject.(otf.User); ok {
			if !user.IsTeamMember(teamID) {
				// User is not a member; refuse access
				return nil, err
			}
		} else {
			// non-user without organization-wide authority; refuse access
			return nil, err
		}
	}

	members, err := a.db.ListTeamMembers(ctx, teamID)
	if err != nil {
		a.Error(err, "listing team members", "team_id", teamID, "subject", subject)
		return nil, err
	}

	a.V(2).Info("listed team members", "team_id", teamID, "subject", subject)

	return members, nil
}

// Registry session services

// CreateRegistrySession creates and persists a registry session.
func (a *app) CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error) {
	return a.createRegistrySession(ctx, organization)
}

// GetRegistrySession retrieves a registry session using a token. Useful for
// checking token is valid.
func (a *app) getRegistrySession(ctx context.Context, token string) (*registrySession, error) {
	// No need for authz because caller is providing an auth token.

	session, err := a.db.getRegistrySession(ctx, token)
	if err != nil {
		a.Error(err, "retrieving registry session", "token", "*****")
		return nil, err
	}

	a.V(2).Info("retrieved registry session", "session", session)

	return session, nil
}

func (a *app) createRegistrySession(ctx context.Context, organization string) (*registrySession, error) {
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

func (a *Application) CreateAgentToken(ctx context.Context, opts otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
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

func (a *Application) ListAgentTokens(ctx context.Context, organization string) ([]*otf.AgentToken, error) {
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

	if err := a.db.DeleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "agent token", at, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted agent token", "agent token", at, "subject", subject)
	return at, nil
}
