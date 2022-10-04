package html

import (
	"context"
	"net/http"
	"time"

	"github.com/leg100/otf"
)

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.render("login.tmpl", w, r, nil)
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := app.DeleteSession(r.Context(), session.Token); err != nil {
		return
	}
	setCookie(w, sessionCookie, session.Token, &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("profile.tmpl", w, r, user)
}

// synchroniseOrganizations ensures an otf user's organization memberships match
// their identity provider account's organization memberships
func synchroniseOrganizations(
	ctx context.Context,
	userService otf.UserService,
	organizationService otf.OrganizationService,
	user *otf.User,
	orgName ...string,
) error {
	var orgs []*otf.Organization

	// Sync orgs
	for _, githubOrganization := range orgName {
		org, err := organizationService.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
			Name: otf.String(githubOrganization),
		})
		if err != nil {
			return err
		}
		orgs = append(orgs, org)
	}
	// A user also gets their own personal organization that matches their
	// username
	org, err := organizationService.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(user.Username()),
	})
	if err != nil {
		return err
	}
	orgs = append(orgs, org)

	// Sync memberships
	if _, err = userService.SyncOrganizationMemberships(ctx, user, orgs); err != nil {
		return err
	}

	return nil
}
