// Package client provides an abstraction for interacting with otf services
// either remotely or locally.
package client

import (
	"context"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/tokens"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/workspace"
)

var (
	_ Client = (*LocalClient)(nil)
	_ Client = (*remoteClient)(nil)
)

type (
	// Client is those service endpoints that support both in-process and remote
	// invocation. Intended for use with the agent (the internal agent is
	// in-process, while the external agent is remote) as well as the CLI.
	Client interface {
		CreateOrganization(ctx context.Context, opts orgcreator.OrganizationCreateOptions) (*organization.Organization, error)
		DeleteOrganization(ctx context.Context, organization string) error

		GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
		GetWorkspaceByName(ctx context.Context, organization, workspace string) (*workspace.Workspace, error)
		ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*workspace.WorkspaceList, error)
		UpdateWorkspace(ctx context.Context, workspaceID string, opts workspace.UpdateOptions) (*workspace.Workspace, error)

		ListVariables(ctx context.Context, workspaceID string) ([]*variable.Variable, error)

		CreateAgentToken(ctx context.Context, opts tokens.CreateAgentTokenOptions) ([]byte, error)
		GetAgentToken(ctx context.Context, token string) (*tokens.AgentToken, error)

		GetPlanFile(ctx context.Context, id string, format run.PlanFormat) ([]byte, error)
		UploadPlanFile(ctx context.Context, id string, plan []byte, format run.PlanFormat) error

		GetLockFile(ctx context.Context, id string) ([]byte, error)
		UploadLockFile(ctx context.Context, id string, lockFile []byte) error

		ListRuns(ctx context.Context, opts run.RunListOptions) (*run.RunList, error)
		GetRun(ctx context.Context, id string) (*run.Run, error)

		StartPhase(ctx context.Context, id string, phase internal.PhaseType, opts run.PhaseStartOptions) (*run.Run, error)
		FinishPhase(ctx context.Context, id string, phase internal.PhaseType, opts run.PhaseFinishOptions) (*run.Run, error)

		DownloadConfig(ctx context.Context, id string) ([]byte, error)

		Watch(context.Context, run.WatchOptions) (<-chan internal.Event, error)

		CreateStateVersion(ctx context.Context, opts state.CreateStateVersionOptions) (*state.Version, error)
		DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error)
		GetCurrentStateVersion(ctx context.Context, workspaceID string) (*state.Version, error)
		RollbackStateVersion(ctx context.Context, svID string) (*state.Version, error)
		DeleteStateVersion(ctx context.Context, svID string) error
		DownloadState(ctx context.Context, svID string) ([]byte, error)
		ListStateVersions(ctx context.Context, options state.StateVersionListOptions) (*state.VersionList, error)

		CreateUser(ctx context.Context, username string, opts ...auth.NewUserOption) (*auth.User, error)
		DeleteUser(ctx context.Context, username string) error
		AddTeamMembership(ctx context.Context, opts auth.TeamMembershipOptions) error
		RemoveTeamMembership(ctx context.Context, opts auth.TeamMembershipOptions) error

		CreateTeam(ctx context.Context, opts auth.CreateTeamOptions) (*auth.Team, error)
		GetTeam(ctx context.Context, organization, team string) (*auth.Team, error)
		DeleteTeam(ctx context.Context, teamID string) error

		Hostname() string

		tokens.RunTokenService
		internal.PutChunkService
		workspace.LockService
	}

	LocalClient struct {
		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		auth.AuthService
		tokens.TokensService
		variable.VariableService
		state.StateService
		workspace.WorkspaceService
		internal.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		logs.LogsService
	}

	remoteClient struct {
		*http.Client
		http.Config

		*stateClient
		*configClient
		*variableClient
		*authClient
		*tokensClient
		*organizationClient
		*organizationCreatorClient
		*workspaceClient
		*runClient
		*logsClient
	}

	stateClient               = state.Client
	configClient              = configversion.Client
	variableClient            = variable.Client
	authClient                = auth.Client
	tokensClient              = tokens.Client
	organizationClient        = organization.Client
	organizationCreatorClient = orgcreator.Client
	workspaceClient           = workspace.Client
	runClient                 = run.Client
	logsClient                = logs.Client
)

// New constructs a client that uses the http to remotely invoke OTF
// services.
func New(config http.Config) (*remoteClient, error) {
	httpClient, err := http.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &remoteClient{
		Client:                    httpClient,
		stateClient:               &stateClient{JSONAPIClient: httpClient},
		configClient:              &configClient{JSONAPIClient: httpClient},
		variableClient:            &variableClient{JSONAPIClient: httpClient},
		authClient:                &authClient{JSONAPIClient: httpClient},
		tokensClient:              &tokensClient{JSONAPIClient: httpClient},
		organizationClient:        &organizationClient{JSONAPIClient: httpClient},
		organizationCreatorClient: &organizationCreatorClient{JSONAPIClient: httpClient},
		workspaceClient:           &workspaceClient{JSONAPIClient: httpClient},
		runClient:                 &runClient{JSONAPIClient: httpClient, Config: config},
		logsClient:                &logsClient{JSONAPIClient: httpClient},
	}, nil
}
