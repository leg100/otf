package forgejo

import (
	"net/http"
	"net/url"

	"github.com/leg100/otf/internal/http/html"
)

// new handler provides a component that includes a form for creating a new
// instance of this provider.
//
// inputs:
// config

func newHandler(r *http.Request) {
	props := newPATProps{
		provider: &VCSProvider{
			Kind:         params.Kind,
			Organization: params.Organization,
		},
	}
	props.scope = "repo (read/write) and user"
	props.tokensURL = url.URL{
		Scheme: "https",
		Host:   h.ForgejoHostname,
		Path:   "/user/settings/applications",
	}
	html.Render(newPAT(props), w, r)
}
