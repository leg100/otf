package agent

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

var (
	ErrInvalidJobStateTransition = errors.New("invalid job state transition")
	ErrMalformedJobSpecString    = errors.New("malformed stringified job spec")
)

type JobStatus string

const (
	JobUnallocated JobStatus = "unallocated"
	JobAllocated   JobStatus = "allocated"
	JobRunning     JobStatus = "running"
	JobFinished    JobStatus = "finished"
	JobErrored     JobStatus = "errored"
	JobCanceled    JobStatus = "canceled"
)

// JobSpec uniquely identifies a job.
type JobSpec struct {
	// ID of the run that this job is for.
	RunID string `json:"run_id"`
	// Phase of run that this job is for.
	Phase internal.PhaseType `json:"phase"`
}

// jobSpecFromString constructs a job spec from a string. The string is
// expected to be in the format run-<id>/<phase>
func jobSpecFromString(spec string) (JobSpec, error) {
	parts := strings.Split(spec, "/")
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "run-") {
		return JobSpec{}, ErrMalformedJobSpecString
	}
	return JobSpec{RunID: parts[0], Phase: internal.PhaseType(parts[1])}, nil
}

func (j JobSpec) String() string {
	return fmt.Sprintf("%s/%s", j.RunID, j.Phase)
}

type Job struct {
	JobSpec
	// Current status of job.
	Status JobStatus
	// Execution mode of job's workspace.
	ExecutionMode workspace.ExecutionMode
	// Name of job's organization
	Organization string
	// ID of job's workspace
	WorkspaceID string
	// ID of agent that this job is allocated to. Only set once job enters
	// JobAllocated state.
	AgentID *string
	// signaled is non-nil when a cancelation signal has been sent to the job
	// and it is true when it has been forceably canceled.
	signaled *bool
}

func newJob(run *otfrun.Run) *Job {
	return &Job{
		JobSpec: JobSpec{
			RunID: run.ID,
			Phase: run.Phase(),
		},
		Status:        JobUnallocated,
		ExecutionMode: run.ExecutionMode,
		Organization:  run.Organization,
		WorkspaceID:   run.WorkspaceID,
	}
}

func (j *Job) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("run_id", j.RunID),
		slog.String("phase", string(j.Phase)),
		slog.String("status", string(j.Status)),
	}
	if j.signaled != nil {
		if *j.signaled {
			attrs = append(attrs, slog.Bool("force_cancel_signal_sent", true))
		} else {
			attrs = append(attrs, slog.Bool("cancel_signal_sent", true))
		}
	}
	return slog.GroupValue(attrs...)
}

func (j *Job) Organizations() []string { return nil }

func (j *Job) IsSiteAdmin() bool   { return true }
func (j *Job) IsOwner(string) bool { return true }

func (j *Job) CanAccessSite(action rbac.Action) bool {
	return false
}

func (j *Job) CanAccessOrganization(action rbac.Action, name string) bool {
	switch action {
	case rbac.GetOrganizationAction, rbac.GetEntitlementsAction, rbac.GetModuleAction, rbac.ListModulesAction:
		return j.Organization == name
	default:
		return false
	}
}

func (j *Job) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	if policy.WorkspaceID != j.WorkspaceID {
		// job is allowed the retrieve the state of *another* workspace only if:
		// (a) workspace is in the same organization as job, or
		// (b) workspace has enabled global remote state (permitting organization-wide
		// state sharing).
		switch action {
		case rbac.GetStateVersionAction, rbac.GetWorkspaceAction, rbac.DownloadStateAction:
			if j.Organization == policy.Organization && policy.GlobalRemoteState {
				return true
			}
		}
		return false
	}
	// allow actions on same workspace as job depending on run phase
	switch action {
	case rbac.DownloadStateAction, rbac.GetStateVersionAction, rbac.GetWorkspaceAction, rbac.GetRunAction, rbac.ListVariableSetsAction, rbac.ListWorkspaceVariablesAction, rbac.PutChunkAction, rbac.DownloadConfigurationVersionAction, rbac.GetPlanFileAction, rbac.CancelRunAction, rbac.ForceCancelRunAction:
		// any phase
		return true
	case rbac.UploadLockFileAction, rbac.UploadPlanFileAction:
		// plan phase
		if j.Phase == internal.PlanPhase {
			return true
		}
	case rbac.GetLockFileAction, rbac.CreateStateVersionAction:
		// apply phase
		if j.Phase == internal.ApplyPhase {
			return true
		}
	}
	return false
}

func (j *Job) CanAccessTeam(rbac.Action, string) bool {
	// Can't access team level actions
	return false
}

func (j *Job) allocate(agentID string) error {
	if j.Status != JobUnallocated {
		return errors.New("job can only be allocated when it is in the unallocated state")
	}
	j.Status = JobAllocated
	j.AgentID = &agentID
	return nil
}

func (j *Job) reallocate(agentID string) error {
	if j.Status != JobAllocated {
		return errors.New("job can only be re-allocated when it is in the allocated state")
	}
	j.AgentID = &agentID
	return nil
}

// cancel job based on current state of its parent run - depending on its state,
// the job is signaled and/or its state is updated too.
func (j *Job) cancel(run *otfrun.Run) (*bool, error) {
	var (
		// whether job be signaled
		signal *bool
	)
	switch run.Status {
	case otfrun.RunPlanning, otfrun.RunApplying:
		if run.CancelSignaledAt != nil {
			// run is still in progress but the user has requested it be
			// canceled, so signal job to gracefully cancel current operation
			signal = internal.Bool(false)
		}
	case otfrun.RunCanceled:
		// run has been canceled so immediately cancel job too unless already
		// canceled
		if j.Status != JobCanceled {
			j.Status = JobCanceled
		}
	case otfrun.RunForceCanceled:
		// run has been forceably canceled, so both signal job to forcefully
		// cancel current operation, and immediately cancel job.
		signal = internal.Bool(true)
		if j.Status != JobCanceled {
			j.Status = JobCanceled
		}
	}
	if signal != nil {
		if j.Status != JobRunning {
			return nil, errors.New("job can only be signaled when in the JobRunning state")
		}
		j.signaled = signal
		return signal, nil
	}
	return nil, nil
}

func (j *Job) updateStatus(to JobStatus) error {
	switch to {
	case JobRunning:
		if j.Status != JobAllocated {
			return ErrInvalidJobStateTransition
		}
	case JobFinished, JobErrored, JobCanceled:
		if j.Status != JobRunning {
			return ErrInvalidJobStateTransition
		}
	default:
		return ErrInvalidJobStateTransition
	}
	j.Status = to
	return nil
}
