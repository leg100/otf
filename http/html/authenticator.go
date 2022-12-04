package html

import (
	"context"
	"net/http"
	"time"

	"github.com/leg100/otf"
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
	NewClient(ctx context.Context, token *oauth2.Token) (otf.CloudClient, error)
	RequestPath() string
	CallbackPath() string
}

func newAuthenticators(app otf.Application, clients []*OAuthClient) ([]*Authenticator, error) {
	var authenticators []*Authenticator
	for _, client := range clients {
		authenticators = append(authenticators, &Authenticator{
			oauthClient: client,
			Application: app,
		})
	}
	return authenticators, nil
}

// exchanging its auth code for a token.
func (a *Authenticator) responseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := a.CallbackHandler(r)
	if err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, loginPath(), http.StatusFound)
		return
	}

	client, err := a.NewClient(r.Context(), token)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := a.synchronise(r.Context(), client)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := createSession(a.Application, w, r, user.ID()); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		setCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, getProfilePath(), http.StatusFound)
	}
}

func (a *Authenticator) synchronise(ctx context.Context, client otf.CloudClient) (*otf.User, error) {
	// give authenticator unlimited access to services
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "authenticator"})

	// Get cloud user
	cuser, err := client.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	// Get otf user; if not exist, create user
	user, err := a.EnsureCreatedUser(ctx, cuser.Username())
	if err != nil {
		return nil, err
	}

	// organizations to be synchronised
	var organizations []*otf.Organization
	// teams to be synchronised
	var teams []*otf.Team

	// Create user's organizations as necessary
	for _, org := range cuser.Organizations() {
		org, err = a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
			Name: otf.String(org.Name()),
		})
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, org)
	}

	// A user also gets their own personal organization matching their username
	personal, err := a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(user.Username()),
	})
	if err != nil {
		return nil, err
	}
	organizations = append(organizations, personal)

	// Create user's teams as necessary
	for _, team := range cuser.Teams() {
		team, err = a.EnsureCreatedTeam(ctx, team.Name(), team.Organization().Name())
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	// And make them an owner of their personal org
	team, err := a.EnsureCreatedTeam(ctx, "owners", personal.Name())
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
