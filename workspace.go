package otf

import (
	"context"
	"time"
)

const (
	RemoteExecutionMode ExecutionMode = "remote"
	LocalExecutionMode  ExecutionMode = "local"
	AgentExecutionMode  ExecutionMode = "agent"
)

type ExecutionMode string

// ExecutionModePtr returns a pointer to an execution mode.
func ExecutionModePtr(m ExecutionMode) *ExecutionMode {
	return &m
}

type Workspace interface {
	ID() string
	CreatedAt() time.Time
	UpdatedAt() time.Time
	String() string
	Name() string
}

// WorkspaceConnector connects a workspace to a VCS repo, subscribing it to
// VCS events that trigger runs.
type WorkspaceConnector interface {
	Connect(ctx context.Context, workspaceID string, opts ConnectWorkspaceOptions) error
	Disconnect(ctx context.Context, workspaceID string) (*Workspace, error)
}

type ConnectWorkspaceOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud host of the repo
}

// CreateWorkspaceOptions represents the options for creating a new workspace.
type CreateWorkspaceOptions struct {
	AllowDestroyPlan           *bool
	AutoApply                  *bool
	Description                *string
	ExecutionMode              *ExecutionMode
	FileTriggersEnabled        *bool
	GlobalRemoteState          *bool
	MigrationEnvironment       *string
	Name                       *string `schema:"name,required"`
	QueueAllRuns               *bool
	SpeculativeEnabled         *bool
	SourceName                 *string
	SourceURL                  *string
	StructuredRunOutputEnabled *bool
	TerraformVersion           *string
	TriggerPrefixes            []string
	WorkingDirectory           *string
	Organization               *string `schema:"organization_name,required"`
	Repo                       *WorkspaceRepo
}
