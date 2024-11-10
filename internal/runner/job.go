package runner

import (
	"errors"
	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	otfrun "github.com/leg100/otf/internal/run"
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

// Job is the unit of work corresponding to a run phase. A job is allocated to
// a runner, which then executes the work through to completion.
type Job struct {
	ID resource.ID `jsonapi:"primary,jobs"`
	// ID of the run that this job is for.
	RunID resource.ID `jsonapi:"attribute" json:"run_id"`
	// Phase of run that this job is for.
	Phase internal.PhaseType `jsonapi:"attribute" json:"phase"`
	// Current status of job.
	Status JobStatus `jsonapi:"attribute" json:"status"`
	// ID of agent pool the job's workspace is assigned to use. If non-nil then
	// the job is allocated to an agent runner belonging to the pool. If nil then
	// the job is allocated to a server runner.
	AgentPoolID *resource.ID `jsonapi:"attribute" json:"agent_pool_id"`
	// Name of job's organization
	Organization string `jsonapi:"attribute" json:"organization"`
	// ID of job's workspace
	WorkspaceID resource.ID `jsonapi:"attribute" json:"workspace_id"`
	// ID of runner that this job is allocated to. Only set once job enters
	// JobAllocated state.
	RunnerID *resource.ID `jsonapi:"attribute" json:"runner_id"`
	// Signaled is non-nil when a cancelation signal has been sent to the job
	// and it is true when it has been forceably canceled.
	Signaled *bool `jsonapi:"attribute" json:"signaled"`
}

func newJob(run *otfrun.Run) *Job {
	return &Job{
		ID:           resource.NewID(resource.JobKind),
		RunID:        run.ID,
		Phase:        run.Phase(),
		Status:       JobUnallocated,
		Organization: run.Organization,
		WorkspaceID:  run.WorkspaceID,
	}
}

func (j *Job) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("run_id", j.RunID.String()),
		slog.String("phase", string(j.Phase)),
		slog.String("status", string(j.Status)),
	}
	if j.Signaled != nil {
		if *j.Signaled {
			attrs = append(attrs, slog.Bool("force_cancel_signal_sent", true))
		} else {
			attrs = append(attrs, slog.Bool("cancel_signal_sent", true))
		}
	}
	return slog.GroupValue(attrs...)
}

func (j *Job) Organizations() []string { return nil }

func (j *Job) IsSiteAdmin() bool   { return false }
func (j *Job) IsOwner(string) bool { return false }
func (j *Job) String() string      { return j.ID.String() }

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

func (j *Job) CanAccessWorkspace(action rbac.Action, policy authz.WorkspacePolicy) bool {
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
	case rbac.DownloadStateAction, rbac.GetStateVersionAction, rbac.GetWorkspaceAction, rbac.GetRunAction, rbac.ListVariableSetsAction, rbac.ListWorkspaceVariablesAction, rbac.PutChunkAction, rbac.DownloadConfigurationVersionAction, rbac.GetPlanFileAction, rbac.CancelRunAction:
		// any phase
		return true
	case rbac.UploadLockFileAction, rbac.UploadPlanFileAction, rbac.ApplyRunAction:
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

func (j *Job) CanAccessTeam(rbac.Action, resource.ID) bool {
	// Can't access team level actions
	return false
}

func (j *Job) allocate(runnerID resource.ID) error {
	if err := j.updateStatus(JobAllocated); err != nil {
		return err
	}
	j.RunnerID = &runnerID
	return nil
}

func (j *Job) reallocate(runnerID resource.ID) error {
	if j.Status != JobAllocated {
		return errors.New("job can only be re-allocated when it is in the allocated state")
	}
	j.RunnerID = &runnerID
	return nil
}

// cancel job based on current state of its parent run - depending on its state,
// the job is signaled and/or its state is updated too.
func (j *Job) cancel(run *otfrun.Run) (*bool, error) {
	// whether job be signaled
	var signal *bool
	switch run.Status {
	case otfrun.RunPlanning, otfrun.RunApplying:
		if run.CancelSignaledAt != nil {
			// run is still in progress but the user has requested it be
			// canceled, so signal job to gracefully cancel current operation
			signal = internal.Bool(false)
		}
	case otfrun.RunCanceled:
		// run has been canceled so immediately cancel job too
		if err := j.updateStatus(JobCanceled); err != nil {
			return nil, err
		}
	case otfrun.RunForceCanceled:
		// run has been forceably canceled, so both signal job to forcefully
		// cancel current operation, and immediately cancel job.
		signal = internal.Bool(true)
		if err := j.updateStatus(JobCanceled); err != nil {
			return nil, err
		}
	}
	if signal != nil {
		if j.Status != JobRunning {
			return nil, errors.New("job can only be signaled when in the JobRunning state")
		}
		j.Signaled = signal
		return signal, nil
	}
	return nil, nil
}

func (j *Job) startJob() error {
	return j.updateStatus(JobRunning)
}

func (j *Job) finishJob(to JobStatus) error {
	return j.updateStatus(to)
}

func (j *Job) updateStatus(to JobStatus) error {
	var isValid bool
	switch j.Status {
	case JobUnallocated:
		switch to {
		case JobAllocated, JobCanceled:
			isValid = true
		}
	case JobAllocated:
		switch to {
		case JobRunning, JobCanceled:
			isValid = true
		}
	case JobRunning:
		switch to {
		case JobFinished, JobCanceled, JobErrored:
			isValid = true
		}
	}
	if isValid {
		j.Status = to
		return nil
	}
	return ErrInvalidJobStateTransition
}
