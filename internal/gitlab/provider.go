package gitlab

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
	var opts ClientOptions
	if err := decode.DecodeConfig(&opts, encodedConfig); err != nil {
		return nil, err
	}
	opts.SkipTLSVerification = transport.TLSClientConfig.InsecureSkipVerify

	// TODO: hostname is currently sourced from the default hostname passed via
	// cli flags, but in future it'll be embedded in encodedConfig, which is
	// sourced from the db or set by a user via the UI.
	opts.Hostname = p.DefaultHostname

	return NewTokenClient(vcs.NewTokenClientOptions(opts))
}

func (p *Provider) CreateFormFields() templ.Component {
	return formFields(formFieldsProps{})
}
