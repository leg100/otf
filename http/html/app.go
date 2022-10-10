package html

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/r3labs/sse/v2"
	"github.com/spf13/pflag"
)

const DefaultPathPrefix = "/"

// Config is the web app configuration.
type Config struct {
	DevMode bool

	cloudConfigs []CloudConfig
}

// NewConfigFromFlags binds flags to the config. The flagset must be parsed
// in order for the config to be populated.
func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}

	cfg.cloudConfigs = append(cfg.cloudConfigs, NewGithubConfigFromFlags(flags))
	cfg.cloudConfigs = append(cfg.cloudConfigs, NewGitlabConfigFromFlags(flags))

	flags.BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")
	return &cfg
}

// Application is the oTF web app.
type Application struct {
	// Static asset server
	staticServer http.FileSystem
	// oTF service accessors
	otf.Application
	// path prefix for all URLs
	pathPrefix string
	// view engine populates and renders templates
	*viewEngine
	// logger for logging messages
	logr.Logger
	// server-side-events server
	*sse.Server
	// enabled authenticators
	authenticators []*Authenticator
}

// AddRoutes adds routes for the html web app.
func AddRoutes(logger logr.Logger, config *Config, services otf.Application, router *otfhttp.Router) error {
	if config.DevMode {
		logger.Info("enabled developer mode")
	}
	views, err := newViewEngine(config.DevMode)
	if err != nil {
		return err
	}

	// Setup SSE server
	sseServer := sse.New()
	// delegate responsibility to sse lib to create/delete streams
	sseServer.AutoStream = true
	// we don't use last-event-item functionality so turn it off
	sseServer.AutoReplay = false

	app := &Application{
		Application:  services,
		staticServer: newStaticServer(config.DevMode),
		pathPrefix:   DefaultPathPrefix,
		viewEngine:   views,
		Logger:       logger,
		Server:       sseServer,
	}

	// Add authenticators for clouds the user has configured
	authenticators, err := NewAuthenticatorsFromConfig(services, config.cloudConfigs...)
	if err != nil {
		return err
	}
	app.authenticators = authenticators

	app.addRoutes(router)
	return nil
}

// AddRoutes adds application routes and middleware to an HTTP multiplexer.
func (app *Application) addRoutes(r *otfhttp.Router) {
	r.Handle("/", http.RedirectHandler("/organizations", http.StatusFound))

	// Static assets (JS, CSS, etc).
	r.PathPrefix("/static/").Handler(http.FileServer(app.staticServer)).Methods("GET")

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	r.StrictSlash(true)

	// routes that don't require authentication.
	r.GET("/login", app.loginHandler)
	for _, auth := range app.authenticators {
		r.GET(auth.RequestPath(), auth.requestHandler)
		r.GET(auth.callbackPath(), auth.responseHandler)
	}

	// routes that require authentication.
	r.Sub(func(r *otfhttp.Router) {
		r.Use((&authMiddleware{app}).authenticate)
		r.Use(setOrganization)

		r.PST("/logout", app.logoutHandler)
		r.GET("/profile", app.profileHandler)
		r.GET("/profile/sessions", app.sessionsHandler)
		r.PST("/profile/sessions/revoke", app.revokeSessionHandler)

		r.GET("/profile/tokens", app.tokensHandler)
		r.PST("/profile/tokens/delete", app.deleteTokenHandler)
		r.GET("/profile/tokens/new", app.newTokenHandler)
		r.PST("/profile/tokens/create", app.createTokenHandler)

		r.GET("/organizations/{organization_name}/agent-tokens", app.listAgentTokens)
		r.PST("/organizations/{organization_name}/agent-tokens/delete", app.deleteAgentToken)
		r.PST("/organizations/{organization_name}/agent-tokens/create", app.createAgentToken)
		r.GET("/organizations/{organization_name}/agent-tokens/new", app.newAgentToken)

		r.GET("/organizations", app.listOrganizations)
		r.GET("/organizations/{organization_name}", app.getOrganization)
		r.GET("/organizations/{organization_name}/edit", app.editOrganization)
		r.PST("/organizations/{organization_name}/update", app.updateOrganization)
		r.PST("/organizations/{organization_name}/delete", app.deleteOrganization)

		r.GET("/organizations/{organization_name}/users", app.listUsers)

		r.GET("/organizations/{organization_name}/teams", app.listTeams)
		r.GET("/organizations/{organization_name}/teams/{team_name}", app.getTeam)
		r.GET("/organizations/{organization_name}/teams/{team_name}/users", app.listTeamUsers)
		r.PST("/organizations/{organization_name}/teams/{team_name}/update", app.updateTeam)

		r.GET("/organizations/{organization_name}/permissions", app.listOrganizationPermissions)

		r.GET("/organizations/{organization_name}/workspaces", app.listWorkspaces)
		r.GET("/organizations/{organization_name}/workspaces/new", app.newWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/create", app.createWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}", app.getWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/edit", app.editWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/update", app.updateWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/delete", app.deleteWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/lock", app.lockWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/unlock", app.unlockWorkspace)

		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/watch", app.watchWorkspace)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs", app.listRuns)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/new", app.newRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/create", app.createRun)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}", app.getRun)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/tail", app.tailRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/delete", app.deleteRun)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/runs/{run_id}/cancel", app.cancelRun)

		// this handles the link the terraform CLI shows during a plan/apply.
		r.GET("/app/{organization_name}/{workspace_name}/runs/{run_id}", app.getRun)

		// terraform login opens a browser to this page
		r.GET("/app/settings/tokens", app.tokensHandler)
	})
}
