package forgejo

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/vcs"
)

type Provider struct {
	DefaultHostname string
}

func (p *Provider) Create(encodedConfig []byte, transport *http.Transport) (vcs.Client, error) {
	var cfg Config
	if err := decode.DecodeConfig(&cfg, encodedConfig); err != nil {
		return nil, err
	}
	cfg.SkipTLSVerification = transport.TLSClientConfig.InsecureSkipVerify

	// TODO: hostname is currently sourced from the default hostname passed via
	// cli flags, but in future it'll be embedded in encodedConfig, which is
	// sourced from the db or set by a user via the UI.
	cfg.Hostname = p.DefaultHostname

	return NewTokenClient(vcs.NewTokenClientOptions(cfg))
}

func (p *Provider) CreateFormFields() templ.Component {
	return formFields(formFieldsProps{})
}
