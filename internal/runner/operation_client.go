package runner

import (
	"context"
	"io"

	"github.com/leg100/otf/internal/resource"
	runpkg "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

// OperationClientCreator creates an OperationClient using the given job token.
type OperationClientCreator func(jobToken string) OperationClient

// OperationClient is a client for an operation to interact with services on
// behalf of its job.
type OperationClient interface {
	GetJob(ctx context.Context, jobID resource.TfeID) (*Job, error)
	GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error)
	Download(ctx context.Context, version string, w io.Writer) (string, error)
	GetRun(ctx context.Context, runID resource.TfeID) (*runpkg.Run, error)
	GetRunPlanFile(ctx context.Context, id resource.TfeID, format runpkg.PlanFormat) ([]byte, error)
	UploadRunPlanFile(ctx context.Context, id resource.TfeID, plan []byte, format runpkg.PlanFormat) error
	GetLockFile(ctx context.Context, id resource.TfeID) ([]byte, error)
	UploadLockFile(ctx context.Context, id resource.TfeID, lockFile []byte) error
	PutChunk(ctx context.Context, opts runpkg.PutChunkOptions) error
	GetWorkspace(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	ListEffectiveVariables(ctx context.Context, runID resource.TfeID) ([]*variable.Variable, error)
	DownloadConfig(ctx context.Context, id resource.TfeID) ([]byte, error)
	CreateStateVersion(ctx context.Context, opts state.CreateStateVersionOptions) (*state.Version, error)
	DownloadCurrentState(ctx context.Context, workspaceID resource.TfeID) ([]byte, error)
	Hostname() string
	GetSSHKeyPrivateKey(ctx context.Context, id resource.TfeID) ([]byte, error)
	awaitJobSignal(ctx context.Context, jobID resource.TfeID) func() (jobSignal, error)
	finishJob(ctx context.Context, jobID resource.TfeID, opts finishJobOptions) error
}

func NewOperationClient(
	Runs runClient,
	Workspaces workspaceClient,
	Variables variablesClient,
	State stateClient,
	Configs configClient,
	Server hostnameClient,
	Jobs operationJobsClient,
	SSHKeys sshKeyClient,
) OperationClient {
	return OperationClient{
		Workspaces: Workspaces,
		Variables:  Variables,
		State:      State,
		Configs:    Configs,
		Runs:       Runs,
		Jobs:       Jobs,
		Server:     Server,
		SSHKeys:    SSHKeys,
	}
}
