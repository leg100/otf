package agent

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

const agentIDHeader = "otf-agent-id"

var (
	_ client = (*InProcClient)(nil)
	_ client = (*rpcClient)(nil)
)

type (
	// client allows the daemon to communicate with the server endpoints.
	client struct {
		runs       runClient
		workspaces workspaceClient
		variables  variablesClient
		//state      stateClient

		//GetPlanFile(ctx context.Context, id string, format run.PlanFormat) ([]byte, error)
		//UploadPlanFile(ctx context.Context, id string, plan []byte, format run.PlanFormat) error
		//GetLockFile(ctx context.Context, id string) ([]byte, error)
		//UploadLockFile(ctx context.Context, id string, lockFile []byte) error
		//DownloadConfig(ctx context.Context, id string) ([]byte, error)
		//CreateStateVersion(ctx context.Context, opts state.CreateStateVersionOptions) (*state.Version, error)
		//DownloadCurrentState(ctx context.Context, workspaceID string) ([]byte, error)
		//Hostname() string

		//internal.PutChunkService
	}

	// InProcClient is a client for in-process communication with the server.
	InProcClient struct {
		tokens.TokensService
		variable.VariableService
		state.StateService
		internal.HostnameService
		configversion.ConfigurationVersionService
		//run.RunService
		workspace.WorkspaceService
		logs.LogsService
		Service
	}

	runClient interface {
		GetRun(ctx context.Context, runID string) (*run.Run, error)
	}

	workspaceClient interface {
		GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
	}

	variablesClient interface {
		ListEffectiveVariables(ctx context.Context, runID string) ([]*variable.Variable, error)
	}
)
