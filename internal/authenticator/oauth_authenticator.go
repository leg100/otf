package authenticator

import (
	"net/http"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/tokens"
)

type (
	// oauthAuthenticator logs people onto the system using an OAuth handshake with an
	// Identity provider before synchronising their user account and various organization
	// and team memberships from the provider.
	oauthAuthenticator struct {
		internal.HostnameService
		tokens.TokensService // for creating session

		oauthClient
	}
)

// ResponseHandler handles exchanging its auth code for a token.
func (a *oauthAuthenticator) ResponseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := a.CallbackHandler(r)
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Login(), http.StatusFound)
		return
	}

	client, err := a.NewClient(r.Context(), token)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// give oauthAuthenticator unlimited access to services
	ctx := internal.AddSubjectToContext(r.Context(), &internal.Superuser{Username: "authenticator"})

	// Get cloud user
	cuser, err := client.GetUser(ctx)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.StartSession(w, r, tokens.StartSessionOptions{
		Username: &cuser.Name,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
