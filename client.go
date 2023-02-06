package otf

import (
	"context"
)

// Client is those services supported by the client application, used in the
// CLI and the agent.
type Client interface {
	CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error)

	CreateWorkspace(ctx context.Context, opts CreateWorkspaceOptions) (*Workspace, error)
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error)
	ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	UpdateWorkspace(ctx context.Context, workspaceID string, opts UpdateWorkspaceOptions) (*Workspace, error)

	ListVariables(ctx context.Context, workspaceID string) ([]Variable, error)

	CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) (*AgentToken, error)
	GetAgentToken(ctx context.Context, token string) (*AgentToken, error)

	GetPlanFile(ctx context.Context, id string, format PlanFormat) ([]byte, error)
	UploadPlanFile(ctx context.Context, id string, plan []byte, format PlanFormat) error
	GetLockFile(ctx context.Context, id string) ([]byte, error)
	UploadLockFile(ctx context.Context, id string, lockFile []byte) error
	ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error)
	GetRun(ctx context.Context, id string) (*Run, error)
	StartPhase(ctx context.Context, id string, phase PhaseType, opts PhaseStartOptions) (*Run, error)
	FinishPhase(ctx context.Context, id string, phase PhaseType, opts PhaseFinishOptions) (*Run, error)

	PutChunk(ctx context.Context, chunk Chunk) error

	DownloadConfig(ctx context.Context, id string) ([]byte, error)

	Watch(context.Context, WatchOptions) (<-chan Event, error)

	CreateRegistrySession(ctx context.Context, organization string) (RegistrySession, error)

	WorkspaceLockService
	StateVersionApp
	HostnameService
}
