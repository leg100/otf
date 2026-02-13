package ui

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

// Handlers registers all UI handlers
type Handlers struct {
	Logger                       logr.Logger
	Runs                         RunService
	Workspaces                   WorkspaceService
	Users                        UserService
	Teams                        TeamClient
	Organizations                OrganizationService
	Modules                      ModuleService
	VCSProviders                 *vcs.Service
	State                        *state.Service
	Runners                      runnerClient
	GithubApp                    GithubAppService
	EngineService                *engine.Service
	Configs                      ConfigVersionService
	HostnameService              HostnameService
	Tokens                       sessionService
	Authorizer                   authz.Interface
	AuthenticatorService         loginService
	VariablesService             *variable.Service
	GithubHostname               *internal.WebURL
	SkipTLSVerification          bool
	SiteToken                    string
	RestrictOrganizationCreation bool

	templates *templates
}

type OrganizationService interface {
	Create(ctx context.Context, opts organization.CreateOptions) (*organization.Organization, error)
	Update(ctx context.Context, name organization.Name, opts organization.UpdateOptions) (*organization.Organization, error)
	Get(ctx context.Context, name organization.Name) (*organization.Organization, error)
	List(ctx context.Context, opts organization.ListOptions) (*resource.Page[*organization.Organization], error)
	Delete(ctx context.Context, name organization.Name) error

	CreateToken(ctx context.Context, opts organization.CreateOrganizationTokenOptions) (*organization.OrganizationToken, []byte, error)
	ListTokens(ctx context.Context, org organization.Name) ([]*organization.OrganizationToken, error)
	DeleteToken(ctx context.Context, org organization.Name) error
}

type ConfigVersionService interface {
	GetSourceIcon(source source.Source) templ.Component
}

type GithubAppService interface {
	CreateApp(context.Context, github.CreateAppOptions) (*github.App, error)
	GetApp(context.Context) (*github.App, error)
	DeleteApp(context.Context) error
	ListInstallations(context.Context) ([]vcs.Installation, error)
	DeleteInstallation(context.Context, int64) error
}

type TeamClient interface {
	Create(ctx context.Context, organization organization.Name, opts team.CreateTeamOptions) (*team.Team, error)
	Get(ctx context.Context, organization organization.Name, teamName string) (*team.Team, error)
	GetByID(ctx context.Context, teamID resource.TfeID) (*team.Team, error)
	List(ctx context.Context, organization organization.Name) ([]*team.Team, error)
	Update(ctx context.Context, teamID resource.TfeID, opts team.UpdateTeamOptions) (*team.Team, error)
	Delete(ctx context.Context, teamID resource.TfeID) error
}

type ModuleService interface {
	GetModuleByID(context.Context, resource.TfeID) (*module.Module, error)
	ListModules(context.Context, module.ListOptions) ([]*module.Module, error)
	ListProviders(context.Context, organization.Name) ([]string, error)
	GetModuleInfo(context.Context, resource.TfeID) (*module.TerraformModule, error)
	PublishModule(context.Context, module.PublishOptions) (*module.Module, error)
	DeleteModule(context.Context, resource.TfeID) (*module.Module, error)
}

type HostnameService interface {
	Hostname() string
	URL(path string) string
	WebhookURL(path string) string
}

type RunService interface {
	Create(context.Context, resource.TfeID, run.CreateOptions) (*run.Run, error)
	List(_ context.Context, opts run.ListOptions) (*resource.Page[*run.Run], error)
	Get(ctx context.Context, id resource.TfeID) (*run.Run, error)
	GetChunk(ctx context.Context, opts run.GetChunkOptions) (run.Chunk, error)
	Cancel(ctx context.Context, id resource.TfeID) error
	ForceCancel(ctx context.Context, id resource.TfeID) error
	Discard(ctx context.Context, id resource.TfeID) error
	Tail(context.Context, run.TailOptions) (<-chan run.Chunk, error)
	Delete(context.Context, resource.TfeID) error
	Apply(context.Context, resource.TfeID) error
	Watch(ctx context.Context) (<-chan pubsub.Event[*run.Event], func())
}

type WorkspaceService interface {
	Get(context.Context, resource.TfeID) (*workspace.Workspace, error)
	Watch(ctx context.Context) (<-chan pubsub.Event[*workspace.Event], func())
	List(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
	ListTags(ctx context.Context, organization organization.Name, opts workspace.ListTagsOptions) (*resource.Page[*workspace.Tag], error)
	Create(ctx context.Context, opts workspace.CreateOptions) (*workspace.Workspace, error)
	GetByName(ctx context.Context, organization organization.Name, workspace string) (*workspace.Workspace, error)
	GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (workspace.Policy, error)
	Update(ctx context.Context, workspaceID resource.TfeID, opts workspace.UpdateOptions) (*workspace.Workspace, error)
	Delete(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	Lock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*workspace.Workspace, error)
	Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*workspace.Workspace, error)
	SetPermission(ctx context.Context, workspaceID, teamID resource.TfeID, role authz.Role) error
	UnsetPermission(ctx context.Context, workspaceID, teamID resource.TfeID) error
	DeleteTags(ctx context.Context, organization organization.Name, tagIDs []resource.TfeID) error
	AddTags(ctx context.Context, workspaceID resource.TfeID, tags []workspace.TagSpec) error
	RemoveTags(ctx context.Context, workspaceID resource.TfeID, tags []workspace.TagSpec) error
}

// runnerClient gives web handlers access to the agents service endpoints
type runnerClient interface {
	CreateAgentPool(ctx context.Context, opts runner.CreateAgentPoolOptions) (*runner.Pool, error)
	UpdateAgentPool(ctx context.Context, poolID resource.TfeID, opts runner.UpdatePoolOptions) (*runner.Pool, error)
	ListAgentPoolsByOrganization(ctx context.Context, organization organization.Name, opts runner.ListPoolOptions) ([]*runner.Pool, error)
	GetAgentPool(ctx context.Context, poolID resource.TfeID) (*runner.Pool, error)
	DeleteAgentPool(ctx context.Context, poolID resource.TfeID) (*runner.Pool, error)

	Register(ctx context.Context, opts runner.RegisterRunnerOptions) (*runner.RunnerMeta, error)
	ListRunners(ctx context.Context, opts runner.ListOptions) ([]*runner.RunnerMeta, error)
	CreateAgentToken(ctx context.Context, poolID resource.TfeID, opts runner.CreateAgentTokenOptions) (*runner.AgentToken, []byte, error)
	GetAgentToken(ctx context.Context, tokenID resource.TfeID) (*runner.AgentToken, error)
	ListAgentTokens(ctx context.Context, poolID resource.TfeID) ([]*runner.AgentToken, error)
	DeleteAgentToken(ctx context.Context, tokenID resource.TfeID) (*runner.AgentToken, error)
}

func NewHandlers(
	Logger logr.Logger,
	Runs *run.Service,
	Workspaces *workspace.Service,
	Users *user.Service,
	Teams *team.Service,
	Organizations *organization.Service,
	Modules *module.Service,
	VCSProviders *vcs.Service,
	State *state.Service,
	Runners *runner.Service,
	GithubApp *github.Service,
	EngineService *engine.Service,
	Configs *configversion.Service,
	HostnameService *internal.HostnameService,
	Tokens *tokens.Service,
	Authorizer *authz.Authorizer,
	AuthenticatorService *authenticator.Service,
	VariablesService *variable.Service,
	GithubHostname *internal.WebURL,
	SkipTLSVerification bool,
	SiteToken string,
	RestrictOrganizationCreation bool,
) *Handlers {
	return &Handlers{
		Logger:                       Logger,
		Runs:                         Runs,
		Workspaces:                   Workspaces,
		Users:                        Users,
		Teams:                        Teams,
		Organizations:                Organizations,
		Modules:                      Modules,
		VCSProviders:                 VCSProviders,
		State:                        State,
		Runners:                      Runners,
		GithubApp:                    GithubApp,
		EngineService:                EngineService,
		Configs:                      Configs,
		HostnameService:              HostnameService,
		Tokens:                       Tokens,
		Authorizer:                   Authorizer,
		AuthenticatorService:         AuthenticatorService,
		VariablesService:             VariablesService,
		GithubHostname:               GithubHostname,
		SkipTLSVerification:          SkipTLSVerification,
		SiteToken:                    SiteToken,
		RestrictOrganizationCreation: RestrictOrganizationCreation,
		templates: &templates{
			Authorizer: Authorizer,
			configs:    Configs,
			workspaces: Workspaces,
			users:      Users,
		},
	}
}

// AddHandlers registers all UI handlers with the router
func (h *Handlers) AddHandlers(r *mux.Router) {
	// Unauthenticated, unprefixed routes
	addLoginHandlers(r, h)

	// Add UI prefix to paths handled by handlers below.
	// NOTE: all UI prefixed paths are intercepted by middleware that mandates
	// the request is authenticated.
	r = r.PathPrefix(paths.UIPrefix).Subrouter()

	addRunHandlers(r, h)
	addTeamHandlers(r, h)
	addUserHandlers(r, h)
	addWorkspaceHandlers(r, h)
	addOrganizationHandlers(r, h)
	addModuleHandlers(r, h)
	addVariableHandlers(r, h)
	addRunnerHandlers(r, h)
	addStateHandlers(r, h)
	addVCSHandlers(r, h)
	addGithubAppHandlers(r, h)
}

func (h *Handlers) renderPage(comp templ.Component, title string, w http.ResponseWriter, r *http.Request, opts ...renderPageOption) {
	props := helpers.LayoutProps{
		Authorizer: h.Authorizer,
		Title:      title,
	}
	for _, o := range opts {
		o(&props)
	}
	// Render component as a child of the layout component.
	layout := helpers.Layout(props)
	html.Render(layout, w, r, html.WithChildren(comp))
}

type renderPageOption func(opts *helpers.LayoutProps)

func withOrganization(org resource.ID) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.Organization = org
	}
}

func withWorkspace(ws *workspace.Workspace) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.Organization = new(ws.Organization)
		opts.Workspace = new(ws.Info())
	}
}

func withBreadcrumbs(crumbs ...helpers.Breadcrumb) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.Breadcrumbs = append(opts.Breadcrumbs, crumbs...)
	}
}

func withContentActions(comp templ.Component) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.ContentActions = comp
	}
}

func withPreContent(comp templ.Component) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.PreContent = comp
	}
}

func withPostContent(comp templ.Component) renderPageOption {
	return func(opts *helpers.LayoutProps) {
		opts.PostContent = comp
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
