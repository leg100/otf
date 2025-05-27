package forgejo

import (
	"net/http"
	"net/url"

	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/vcs"
)

// new handler provides a component that includes a form for creating a new
// instance of this provider.
//
// inputs:
// config

func newHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
		Kind         vcs.Kind          `schema:"kind,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

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
