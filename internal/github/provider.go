package github

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/vcs"
)

// TokenProvider creates a github client using a personal access token.
type TokenProvider struct {
	DefaultHostname string
}

func (p *TokenProvider) Create(encodedConfig []byte, transport *http.Transport) (vcs.Client, error) {
	var opts ClientOptions
	if err := decode.DecodeConfig(&opts, encodedConfig); err != nil {
		return nil, err
	}
	opts.SkipTLSVerification = transport.TLSClientConfig.InsecureSkipVerify

	// TODO: hostname is currently sourced from the default hostname passed via
	// cli flags, but in future it'll be embedded in encodedConfig, which is
	// sourced from the db or set by a user via the UI.
	opts.Hostname = p.DefaultHostname

	return NewClient(opts)
}

func (p *TokenProvider) CreateFormFields() templ.Component {
	return tokenFormFields(tokenFormFieldsProps{})
}

// AppProvider creates a github client using a github app install.
type AppProvider struct {
	DefaultHostname string
	Service         *Service
}

func (p *AppProvider) Create(encodedConfig []byte, transport *http.Transport) (vcs.Client, error) {
	var opts ClientOptions
	if err := decode.DecodeConfig(&opts, encodedConfig); err != nil {
		return nil, err
	}
	opts.SkipTLSVerification = transport.TLSClientConfig.InsecureSkipVerify

	creds, err := p.Service.GetInstallCredentials(context.Background(), opts.InstallCredentials.ID)
	if err != nil {
		return nil, err
	}
	opts.InstallCredentials = creds

	// TODO: hostname is currently sourced from the default hostname passed via
	// cli flags, but in future it'll be embedded in encodedConfig, which is
	// sourced from the db or set by a user via the UI.
	opts.Hostname = p.DefaultHostname

	return NewClient(opts)
}

func (p *AppProvider) CreateFormFields() templ.Component {
	return appFormFields(appFormFieldsProps{})
}
