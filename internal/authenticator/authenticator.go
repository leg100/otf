// Package authenticator is responsible for handling the authentication of users with
// third party identity providers.
package authenticator

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/tokens"
)

type (
	authenticator interface {
		RequestPath() string
		CallbackPath() string
		RequestHandler(w http.ResponseWriter, r *http.Request)
		ResponseHandler(w http.ResponseWriter, r *http.Request)
	}

	service struct {
		renderer       html.Renderer
		authenticators []authenticator
	}

	Options struct {
		logr.Logger
		html.Renderer

		internal.HostnameService
		organization.OrganizationService
		auth.AuthService
		tokens.TokensService

		Configs     []cloud.CloudOAuthConfig
		OIDCConfigs []cloud.OIDCConfig
	}
)
