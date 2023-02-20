package otf

import (
	"context"
	"time"
)

func (r RunStatus) String() string { return string(r) }

const (
	// List all available run statuses supported in OTF.
	RunApplied            RunStatus = "applied"
	RunApplyQueued        RunStatus = "apply_queued"
	RunApplying           RunStatus = "applying"
	RunCanceled           RunStatus = "canceled"
	RunForceCanceled      RunStatus = "force_canceled"
	RunConfirmed          RunStatus = "confirmed"
	RunDiscarded          RunStatus = "discarded"
	RunErrored            RunStatus = "errored"
	RunPending            RunStatus = "pending"
	RunPlanQueued         RunStatus = "plan_queued"
	RunPlanned            RunStatus = "planned"
	RunPlannedAndFinished RunStatus = "planned_and_finished"
	RunPlanning           RunStatus = "planning"

	PlanFormatBinary = "bin"  // plan file in binary format
	PlanFormatJSON   = "json" // plan file in json format
)

var (
	ActiveRun = []RunStatus{
		RunApplyQueued,
		RunApplying,
		RunConfirmed,
		RunPlanQueued,
		RunPlanned,
		RunPlanning,
	}
	IncompleteRun = append(ActiveRun, RunPending)
	CompletedRun  = []RunStatus{
		RunApplied,
		RunErrored,
		RunDiscarded,
		RunCanceled,
		RunForceCanceled,
	}
)

// RunStatus represents a run state.
type RunStatus string

// PlanFormat is the format of the plan file
type PlanFormat string

// Run value type for use outside the run package.
type Run struct {
	ID                     string
	CreatedAt              time.Time
	String                 string
	IsDestroy              bool
	ForceCancelAvailableAt *time.Time
	Message                string
	Organization           string
	Phase                  PhaseType
	Refresh                bool
	RefreshOnly            bool
	ReplaceAddrs           []string
	TargetAddrs            []string
	AutoApply              bool
	Speculative            bool
	Status                 RunStatus
	WorkspaceID            string
	ConfigurationVersionID string
	HasChanges             bool
	Latest                 bool
	Plan                   Plan
	Apply                  Apply
}

type (
	Plan  struct{}
	Apply struct{}
)

// RunList represents a list of runs.
type RunList struct {
	*Pagination
	Items []Run
}

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	ListOptions
	// Filter by run statuses (with an implicit OR condition)
	Statuses []RunStatus `schema:"statuses,omitempty"`
	// Filter by workspace ID
	WorkspaceID *string `schema:"workspace_id,omitempty"`
	// Filter by organization name
	Organization *string `schema:"organization_name,omitempty"`
	// Filter by workspace name
	WorkspaceName *string `schema:"workspace_name,omitempty"`
	// Filter by speculative or non-speculative
	Speculative *bool `schema:"-"`
	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include,omitempty"`
}

//type (
//	Plan interface {
//		ResourceReport() *ResourceReport
//	}
//	Apply interface {
//		ResourceReport() *ResourceReport
//	}
//)
//
type RunDB interface {
	GetRun(context.Context, string) (Run, error)
}

// RunService implementations allow interactions with runs
type RunService interface {
	// Create a new run with the given options.
	// CreateRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error)
	// Get retrieves a run with the given ID.
	GetRun(ctx context.Context, id string) (Run, error)
	// List lists runs according to the given options.
	//ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error)
	// Delete deletes a run with the given ID.
	//DeleteRun(ctx context.Context, id string) error
	// EnqueuePlan enqueues a plan
	//EnqueuePlan(ctx context.Context, id string) (*Run, error)
	// Apply a run with the given ID.
	//
	// TODO: return run
	//ApplyRun(ctx context.Context, id string, opts RunApplyOptions) error
	// Discard discards a run with the given ID.
	//
	// TODO: return run
	//DiscardRun(ctx context.Context, id string, opts RunDiscardOptions) error
	// Cancel run.
	//
	// TODO: return run
	//CancelRun(ctx context.Context, id string, opts RunCancelOptions) error
	// Forcefully cancel a run.
	//
	// TODO: return run
	//ForceCancelRun(ctx context.Context, id string, opts RunForceCancelOptions) error
	// Start a run phase.
	//StartPhase(ctx context.Context, id string, phase PhaseType, opts PhaseStartOptions) (*Run, error)
	// Finish a run phase.
	//FinishPhase(ctx context.Context, id string, phase PhaseType, opts PhaseFinishOptions) (*Run, error)
	// GetPlanFile retrieves a run's plan file with the requested format.
	//GetPlanFile(ctx context.Context, id string, format PlanFormat) ([]byte, error)
	// UploadPlanFile saves a run's plan file with the requested format.
	//UploadPlanFile(ctx context.Context, id string, plan []byte, format PlanFormat) error
	// GetLockFile retrieves a run's lock file (.terraform.lock.hcl)
	//GetLockFile(ctx context.Context, id string) ([]byte, error)
	// UploadLockFile saves a run's lock file (.terraform.lock.hcl)
	//UploadLockFile(ctx context.Context, id string, lockFile []byte) error
	// StartRun creates and starts a run.
	//StartRun(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*Run, error)
}
