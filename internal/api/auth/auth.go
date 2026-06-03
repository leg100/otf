package auth

import (
	"net/http"
	"strings"

	"github.com/leg100/otf/internal/authn"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/tfeapi"
)

var _ authn.AuthenticationRoute = (*Route)(nil)

// Route is the authentication route for the API
type Route struct{}

func (a *Route) IsPath(path string) bool {
	return strings.HasPrefix(path, tfeapi.APIPrefixV2) ||
		strings.HasPrefix(path, tfeapi.ModuleV1Prefix) ||
		strings.HasPrefix(path, otfhttp.APIBasePath)
}

func (a *Route) HandleError(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}
