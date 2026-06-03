package auth

import (
	"net/http"
	"strings"

	"github.com/leg100/otf/internal/authn"
	uipath "github.com/leg100/otf/internal/path"
	"github.com/leg100/otf/internal/ui/helpers"
)

var _ authn.AuthenticationRoute = (*Route)(nil)

// Route is the authentication route for the UI
type Route struct{}

func (a *Route) IsPath(path string) bool {
	return strings.HasPrefix(path, uipath.Prefix)
}

func (a *Route) HandleError(err error, w http.ResponseWriter, r *http.Request) {
	helpers.FlashError(w, err.Error())
	helpers.SendUserToLoginPage(w, r)
}
