package html

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

// githubLogin is called upon a successful Github login. A new user is created
// if they don't already exist.
func (app *Application) githubLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// If the OAuth handshake returns an error, return the user to the login
	// page along with a flash alert.
	token, err := app.oauth.responseHandler(r)
	if err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, loginPath(), http.StatusFound)
		return
	}

	client, err := app.oauth.newClient(ctx, token)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	githubOrganizations, err := client.ListOrganizations(ctx, "")
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(githubOrganizations) == 0 {
		flashError(w, "no github organizations found")
		http.Redirect(w, r, loginPath(), http.StatusFound)
		return
	}

	guser, err := client.GetUser(ctx, "")
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get named user; if not exist create user
	user, err := app.UserService().EnsureCreated(ctx, *guser.Login)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = synchroniseOrganizations(ctx, app.UserService(), app.OrganizationService(), user, githubOrganizations...)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create session data
	data, err := otf.NewSessionData(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create session and redirect user to their profile
	session, err := app.UserService().CreateSession(r.Context(), user, data)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setCookie(w, sessionCookie, session.Token, &session.Expiry)
	http.Redirect(w, r, getProfilePath(), http.StatusFound)
}

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.render("login.tmpl", w, r, nil)
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getCtxSession(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := app.UserService().DeleteSession(r.Context(), session.Token); err != nil {
		return
	}
	setCookie(w, sessionCookie, session.Token, &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getCtxUser(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("profile.tmpl", w, r, user)
}

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getCtxUser(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("session_list.tmpl", w, r, user)
}

func (app *Application) newTokenHandler(w http.ResponseWriter, r *http.Request) {
	app.render("token_new.tmpl", w, r, nil)
}

func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getCtxUser(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var opts otf.TokenCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	token, err := app.UserService().CreateToken(r.Context(), user, &opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "created token: "+token.Token())
	http.Redirect(w, r, listTokenPath(), http.StatusFound)
}

func (app *Application) tokensHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getCtxUser(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("token_list.tmpl", w, r, user)
}

func (app *Application) deleteTokenHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getCtxUser(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := app.UserService().DeleteToken(r.Context(), user, id); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "Deleted token")
	http.Redirect(w, r, listTokenPath(), http.StatusFound)
}

func (app *Application) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		writeError(w, "missing token", http.StatusUnprocessableEntity)
		return
	}
	if err := app.UserService().DeleteSession(r.Context(), token); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "Revoked session")
	http.Redirect(w, r, listSessionPath(), http.StatusFound)
}

// synchroniseOrganizations ensures an otf user's organization memberships match
// their github user's organization memberships
func synchroniseOrganizations(
	ctx context.Context,
	userService otf.UserService,
	organizationService otf.OrganizationService,
	user *otf.User,
	githubOrganization ...*github.Organization) error {

	var orgs []*otf.Organization

	for _, githubOrganization := range githubOrganization {
		org, err := organizationService.EnsureCreated(ctx, otf.OrganizationCreateOptions{
			Name: githubOrganization.Login,
		})
		if err != nil {
			return err
		}
		orgs = append(orgs, org)
	}

	_, err := userService.SyncOrganizationMemberships(ctx, user, orgs)
	if err != nil {
		return err
	}

	return nil
}
