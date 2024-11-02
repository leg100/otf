package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

var ErrCannotDeleteOnlyOwner = errors.New("cannot remove the last owner")

type (
	Service struct {
		logr.Logger

		site         internal.Authorizer // authorizes site access
		organization internal.Authorizer // authorizes org access
		teams        *team.Service

		db     *pgdb
		web    *webHandlers
		tfeapi *tfe
		api    *api

		*userTokenFactory
	}

	Options struct {
		SiteToken     string
		TokensService *tokens.Service
		TeamService   *team.Service

		*sql.DB
		*tfeapi.Responder
		html.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
		db:           &pgdb{opts.DB, opts.Logger},
		userTokenFactory: &userTokenFactory{
			tokens: opts.TokensService,
		},
		teams: opts.TeamService,
	}
	svc.web = &webHandlers{
		Renderer:  opts.Renderer,
		teams:     opts.TeamService,
		tokens:    opts.TokensService,
		siteToken: opts.SiteToken,
		users:     &svc,
	}
	svc.tfeapi = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}

	// Whenever an owners team is created, add the creator as a member.
	opts.TeamService.AfterCreateTeam(func(ctx context.Context, team *team.Team) error {
		if team.Name != "owners" {
			return nil
		}
		user, err := UserFromContext(ctx)
		if err != nil {
			return fmt.Errorf("adding owner to owners team: %w", err)
		}
		if err := svc.AddTeamMembership(ctx, team.ID, []string{user.Username}); err != nil {
			return fmt.Errorf("adding owner to owners team: %w", err)
		}
		// and add team to the context user too so that they can immediately
		// enjoy the newly conferred owner privileges on successive calls.
		user.Teams = append(user.Teams, team)
		return nil
	})
	// Fetch users when API calls request users be included in the
	// response
	opts.Register(tfeapi.IncludeUsers, svc.tfeapi.includeUsers)
	// Register site token and site admin with the auth middleware, to permit
	// the latter to authenticate using the former.
	opts.TokensService.RegisterSiteToken(opts.SiteToken, &SiteAdmin)
	// Register with auth middleware the user token kind and a means of
	// retrieving user corresponding to token.
	opts.TokensService.RegisterKind(UserTokenKind, func(ctx context.Context, tokenID resource.ID) (internal.Subject, error) {
		return svc.GetUser(ctx, UserSpec{AuthenticationTokenID: internal.String(tokenID)})
	})
	// Register with auth middleware the ability to get or create a user given a
	// username.
	opts.TokensService.RegisterUISubjectGetterOrCreator(func(ctx context.Context, username string) (internal.Subject, error) {
		user, err := svc.GetUser(ctx, UserSpec{Username: &username})
		if err == internal.ErrResourceNotFound {
			user, err = svc.Create(ctx, username)
		}
		return user, err
	})

	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.tfeapi.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *Service) Create(ctx context.Context, username string, opts ...NewUserOption) (*User, error) {
	subject, err := a.site.CanAccess(ctx, rbac.CreateUserAction, "")
	if err != nil {
		return nil, err
	}

	user := NewUser(username, opts...)

	if err := a.db.CreateUser(ctx, user); err != nil {
		a.Error(err, "creating user", "username", username, "subject", subject)
		return nil, err
	}

	a.V(0).Info("created user", "username", username, "subject", subject)

	return user, nil
}

func (a *Service) GetUser(ctx context.Context, spec UserSpec) (*User, error) {
	subject, err := a.site.CanAccess(ctx, rbac.GetUserAction, "")
	if err != nil {
		return nil, err
	}

	user, err := a.db.getUser(ctx, spec)
	if err != nil {
		a.Error(err, "retrieving user", "spec", spec, "subject", subject)
		return nil, err
	}

	a.V(9).Info("retrieved user", "username", user.Username, "subject", subject)

	return user, nil
}

// List lists all users.
func (a *Service) List(ctx context.Context) ([]*User, error) {
	_, err := a.site.CanAccess(ctx, rbac.ListUsersAction, "")
	if err != nil {
		return nil, err
	}

	return a.db.listUsers(ctx)
}

// ListOrganizationUsers lists an organization's users
func (a *Service) ListOrganizationUsers(ctx context.Context, organization string) ([]*User, error) {
	_, err := a.organization.CanAccess(ctx, rbac.ListUsersAction, organization)
	if err != nil {
		return nil, err
	}

	return a.db.listOrganizationUsers(ctx, organization)
}

// ListTeamUsers lists users that are members of the given team. The caller
// needs either organization-wide authority to call this endpoint, or they need
// to be a member of the team.
func (a *Service) ListTeamUsers(ctx context.Context, teamID resource.ID) ([]*User, error) {
	team, err := a.teams.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.ListUsersAction, team.Organization)
	if err != nil {
		return nil, err
	}

	members, err := a.db.listTeamUsers(ctx, teamID)
	if err != nil {
		a.Error(err, "listing team members", "team_id", teamID, "subject", subject)
		return nil, err
	}

	a.V(9).Info("listed team members", "team_id", teamID, "subject", subject)

	return members, nil
}

func (a *Service) Delete(ctx context.Context, username string) error {
	subject, err := a.site.CanAccess(ctx, rbac.DeleteUserAction, "")
	if err != nil {
		return err
	}

	err = a.db.DeleteUser(ctx, UserSpec{Username: internal.String(username)})
	if err != nil {
		a.Error(err, "deleting user", "username", username, "subject", subject)
		return err
	}

	a.V(2).Info("deleted user", "username", username, "subject", subject)

	return nil
}

// AddTeamMembership adds users to a team. If a user does not exist then the
// user is created first.
func (a *Service) AddTeamMembership(ctx context.Context, teamID resource.ID, usernames []string) error {
	team, err := a.teams.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.AddTeamMembershipAction, team.Organization)
	if err != nil {
		return err
	}

	err = a.db.Tx(ctx, func(ctx context.Context, _ *sqlc.Queries) error {
		// Check each username: if user does not exist then create user.
		for _, username := range usernames {
			_, err := a.db.getUser(ctx, UserSpec{Username: &username})
			if errors.Is(err, internal.ErrResourceNotFound) {
				if _, err := a.Create(ctx, username); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		if err := a.db.addTeamMembership(ctx, teamID, usernames...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		a.Error(err, "adding team membership", "user", usernames, "team", teamID, "subject", subject)
		return err
	}

	a.V(0).Info("added team membership", "users", usernames, "team", teamID, "subject", subject)

	return nil
}

// RemoveTeamMembership removes users from a team.
func (a *Service) RemoveTeamMembership(ctx context.Context, teamID resource.ID, usernames []string) error {
	team, err := a.teams.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.RemoveTeamMembershipAction, team.Organization)
	if err != nil {
		return err
	}

	// check whether *all* members of the owners group are going to be removed
	// (which is not allowed)
	if team.Name == "owners" {
		if owners, err := a.ListTeamUsers(ctx, team.ID); err != nil {
			a.Error(err, "removing team membership: listing team members", "team_id", team.ID, "subject", subject)
			return err
		} else if len(owners) <= len(usernames) {
			return ErrCannotDeleteOnlyOwner
		}
	}

	if err := a.db.removeTeamMembership(ctx, teamID, usernames...); err != nil {
		a.Error(err, "removing team membership", "users", usernames, "team", teamID, "subject", subject)
		return err
	}
	a.V(0).Info("removed team membership", "users", usernames, "team", teamID, "subject", subject)

	return nil
}

// SetSiteAdmins authoritatively promotes users with the given usernames to site
// admins. If no such users exist then they are created. Any unspecified users
// that are currently site admins are demoted.
func (a *Service) SetSiteAdmins(ctx context.Context, usernames ...string) error {
	for _, username := range usernames {
		_, err := a.db.getUser(ctx, UserSpec{Username: &username})
		if err == internal.ErrResourceNotFound {
			if _, err = a.Create(ctx, username); err != nil {
				return err
			}
		}
	}
	promoted, demoted, err := a.db.setSiteAdmins(ctx, usernames...)
	if err != nil {
		a.Error(err, "setting site admins", "users", usernames)
		return err
	}
	a.V(0).Info("set site admins", "admins", usernames, "promoted", promoted, "demoted", demoted)
	return nil
}

// User API token endpoints

// CreateToken creates a user token. Only users can create a user token, and
// they can only create a token for themselves.
func (a *Service) CreateToken(ctx context.Context, opts CreateUserTokenOptions) (*UserToken, []byte, error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	ut, token, err := a.NewUserToken(user.Username, opts)
	if err != nil {
		a.Error(err, "constructing user token", "user", user)
		return nil, nil, err
	}

	if err := a.db.createUserToken(ctx, ut); err != nil {
		a.Error(err, "creating token", "user", user)
		return nil, nil, err
	}

	a.V(1).Info("created user token", "user", user)

	return ut, token, nil
}

func (a *Service) ListTokens(ctx context.Context) ([]*UserToken, error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return a.db.listUserTokens(ctx, user.Username)
}

func (a *Service) DeleteToken(ctx context.Context, tokenID resource.ID) error {
	user, err := UserFromContext(ctx)
	if err != nil {
		return err
	}

	token, err := a.db.getUserToken(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving token", "user", user)
		return err
	}

	if user.Username != token.Username {
		return internal.ErrAccessNotPermitted
	}

	if err := a.db.deleteUserToken(ctx, tokenID); err != nil {
		a.Error(err, "deleting user token", "user", user)
		return err
	}

	a.V(1).Info("deleted user token", "username", user)

	return nil
}
