package forgejo

import (
	"net/http"
	"net/url"

	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/vcs"
)

type Service struct {
	hostname string
}

type NewClientOptions struct {
	Token               string
	Hostname            string
	SkipTLSVerification bool
}

func (s *Service) NewClient(opts NewClientOptions) (vcs.Client, error) {
	return nil, nil
}

func (s *Service) NewHandler(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	html.Render(new(
}

type NewComponentOptions struct {
	Token               string
	Hostname            string
	SkipTLSVerification bool
}
