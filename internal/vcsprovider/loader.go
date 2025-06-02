package vcsprovider

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/vcs"
)

type ClientCreator interface {
	// Create a vcs client from config. The config is encoded as either JSON or as
	// a URL query string. The implementation is advised to use DecodeConfig to
	// decode the config.
	Create(config []byte, transport *http.Transport) (vcs.Client, error)
	// CreateFormFields provides the fields for an html form with which to create a
	// vcs.Client.
	CreateFormFields() templ.Component
}
