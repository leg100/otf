package html

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html/paths"
	"golang.org/x/oauth2"
)

// Authenticator logs people onto the system using an OAuth handshake with an
// Identity provider before synchronising their user account and various organization
// and team memberships from the provider.
type Authenticator struct {
	otf.Application
	oauthClient
}

// oauthClient is an oauth client for the authenticator, implemented as an
// interface to permit swapping out for testing purposes.
type oauthClient interface {
	RequestHandler(w http.ResponseWriter, r *http.Request)
	CallbackHandler(*http.Request) (*oauth2.Token, error)
	NewClient(ctx context.Context, token *oauth2.Token) (cloud.Client, error)
	RequestPath() string
	CallbackPath() string
	String() string
}

func newAuthenticators(logger logr.Logger, app otf.Application, configs []*cloud.CloudOAuthConfig) ([]*Authenticator, error) {
	var authenticators []*Authenticator

	for _, cfg := range configs {
		if cfg.OAuthConfig.ClientID == "" && cfg.OAuthConfig.ClientSecret == "" {
			// skip creating oauth client when creds are unspecified
			continue
		}
		client, err := NewOAuthClient(NewOAuthClientConfig{
			CloudOAuthConfig: cfg,
			hostname:         app.Hostname(),
		})
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, &Authenticator{
			oauthClient: client,
			Application: app,
		})
		logger.V(2).Info("activated oauth client", "name", cfg, "hostname", cfg.Hostname)
	}
	return authenticators, nil
}

// exchanging its auth code for a token.
func (a *Authenticator) responseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := a.CallbackHandler(r)
	if err != nil {
		FlashError(w, err.Error())
		http.Redirect(w, r, paths.Login(), http.StatusFound)
		return
	}

	client, err := a.NewClient(r.Context(), token)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := a.synchronise(r.Context(), client)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := createSession(a.Application, w, r, user.ID()); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		setCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, paths.Profile(), http.StatusFound)
	}
}

func (a *Authenticator) synchronise(ctx context.Context, client cloud.Client) (*otf.User, error) {
	// give authenticator unlimited access to services
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "authenticator"})

	// Get cloud user
	cuser, err := client.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	// Get otf user; if not exist, create user
	user, err := a.EnsureCreatedUser(ctx, cuser.Name)
	if err != nil {
		return nil, err
	}

	// organization names to be synchronised
	var organizations []string
	// teams to be synchronised
	var teams []*otf.Team

	// Create organization for each cloud organization
	for _, corg := range cuser.Organizations {
		org, err := a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
			Name: otf.String(corg),
		})
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, org.Name())
	}

	// A user also gets their own personal organization matching their username
	personal, err := a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(user.Username()),
	})
	if err != nil {
		return nil, err
	}
	organizations = append(organizations, personal.Name())

	// Create team for each cloud team
	for _, cteam := range cuser.Teams {
		team, err := a.EnsureCreatedTeam(ctx, otf.CreateTeamOptions{
			Name:         cteam.Name,
			Organization: cteam.Organization,
		})
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	// And make them an owner of their personal org
	team, err := a.EnsureCreatedTeam(ctx, otf.CreateTeamOptions{
		Name:         "owners",
		Organization: personal.Name(),
	})
	if err != nil {
		return nil, err
	}
	teams = append(teams, team)

	// Synchronise user's memberships so that they match those of the cloud
	// user.
	if _, err = a.SyncUserMemberships(ctx, user, organizations, teams); err != nil {
		return nil, err
	}

	return user, nil
}
