// Package authenticator is responsible for handling the authentication of users with
// third party identity providers.
package authenticator

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
	"github.com/leg100/otf/tokens"
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

		otf.HostnameService
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		auth.AuthService
		tokens.TokensService

		Configs     []cloud.CloudOAuthConfig
		OIDCConfigs []cloud.OIDCConfig
	}
)
