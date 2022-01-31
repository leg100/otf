package html

import (
	"context"
	"errors"
	"net/http"
	"path"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
)

type GithubClient interface {
	GetUser(ctx context.Context, name string) (*github.User, error)
	ListOrganizations(ctx context.Context, name string) ([]*github.Organization, error)
}

var (
	ErrNoGithubOrganizationsFound = errors.New("no github organizations found")
)

// githubLogin is called upon a successful Github login. A new user is created
// if they don't already exist.
func (app *Application) githubLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// If the OAuth handshake returns an error, return the user to the login
	// page along with a flash alert.
	token, err := app.oauth.responseHandler(r)
	if err != nil {
		app.sessions.FlashError(r, err.Error())
		http.Redirect(w, r, app.route("login"), http.StatusFound)
		return
	}

	client, err := app.oauth.newClient(ctx, token)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
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

	// Fetch their github organization memberships and ensure that each github
	// organization has a corresponding oTF organization (if not, create it) and
	// then update the user with their corresponding oTF organization
	// memberships.

	githubOrganizations, err := client.ListOrganizations(ctx, "")
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(githubOrganizations) == 0 {
		app.sessions.FlashError(r, "no github organizations found")
		http.Redirect(w, r, app.route("login"), http.StatusFound)
		return
	}

	err = synchroniseOrganizations(ctx, app.UserService(), app.OrganizationService(), user, githubOrganizations...)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transfer session from anonymous to named user.
	if err = app.sessions.TransferSession(ctx, user); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, app.route("getProfile"), http.StatusFound)
}

// requireAuthentication is middleware that insists on the user being
// authenticated before passing on the request.
func (app *Application) requireAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !app.sessions.IsAuthenticated(r.Context()) {
			http.Redirect(w, r, app.route("login"), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	tdata := app.newTemplateData(r, nil)

	if err := app.renderTemplate("login.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.sessions.Destroy(r.Context(), w); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func (app *Application) meHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, path.Join(r.URL.Path, "profile"), http.StatusFound)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	user := app.sessions.getUserFromContext(r.Context())

	tdata := app.newTemplateData(r, user)

	if err := app.renderTemplate("profile.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.sessions.getUserFromContext(r.Context())

	tdata := app.newTemplateData(r, user)

	if err := app.renderTemplate("session_list.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) newTokenHandler(w http.ResponseWriter, r *http.Request) {
	tdata := app.newTemplateData(r, nil)

	if err := app.renderTemplate("token_new.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	user := app.sessions.getUserFromContext(r.Context())

	var opts otf.TokenCreateOptions
	if err := decodeAll(r, &opts); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := app.UserService().CreateToken(r.Context(), user.User, &opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		app.sessions.FlashSuccess(r, "created token: ", token.Token)
	}

	http.Redirect(w, r, app.route("listToken"), http.StatusFound)
}

func (app *Application) tokensHandler(w http.ResponseWriter, r *http.Request) {
	user := app.sessions.getUserFromContext(r.Context())

	tdata := app.newTemplateData(r, user)

	if err := app.renderTemplate("token_list.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) deleteTokenHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "missing id", http.StatusUnprocessableEntity)
		return
	}

	if err := app.UserService().DeleteToken(r.Context(), id); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.sessions.FlashSuccess(r, "Deleted token"); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, app.route("listToken"), http.StatusFound)
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

	if err := app.sessions.FlashSuccess(r, "Revoked session"); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, app.route("listSession"), http.StatusFound)
}

// synchroniseOrganizations ensures an otf user's organization memberships match
// their github user's organization memberships
func synchroniseOrganizations(
	ctx context.Context,
	userService otf.UserService,
	organizationService otf.OrganizationService,
	user *otf.User,
	githubOrganization ...*github.Organization) error {

	user.Organizations = []*otf.Organization{}

	for _, githubOrganization := range githubOrganization {
		org, err := organizationService.EnsureCreated(ctx, otf.OrganizationCreateOptions{
			Name: githubOrganization.Login,
		})
		if err != nil {
			return err
		}
		user.Organizations = append(user.Organizations, org)
	}

	if err := userService.Update(ctx, user.Username, user); err != nil {
		return err
	}

	return nil
}
