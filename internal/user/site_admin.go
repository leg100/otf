package user

import (
	"net/http"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

var (
	SiteAdminUsername = Username{name: "site-admin"}
	// SiteAdminID is the hardcoded user id for the site admin user. The ID must
	// be the same as the hardcoded value in the database migrations.
	SiteAdminID = resource.MustHardcodeTfeID(resource.UserKind, "36atQC2oGQng7pVz")
	SiteAdmin   = User{ID: SiteAdminID, Username: SiteAdminUsername}
)

// SiteAdminAuthenticator authenticates API requests from the site admin user.
type SiteAdminAuthenticator struct {
	SiteToken string
}

func (a *SiteAdminAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		return nil, nil
	}
	token, err := tokens.ParseBearerToken(bearer)
	if err != nil {
		return nil, err
	}
	if a.SiteToken == token {
		// Authenticated as site admin
		return &SiteAdmin, nil
	}
	// Not a site admin auth request.
	return nil, nil
}
