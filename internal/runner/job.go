package runner

import (
	"errors"
	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
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
	ID resource.TfeID `jsonapi:"primary,jobs"`
	// ID of the run that this job is for.
	RunID resource.TfeID `jsonapi:"attribute" json:"run_id"`
	// Phase of run that this job is for.
	Phase internal.PhaseType `jsonapi:"attribute" json:"phase"`
	// Current status of job.
	Status JobStatus `jsonapi:"attribute" json:"status"`
	// ID of agent pool the job's workspace is assigned to use. If non-nil then
	// the job is allocated to an agent runner belonging to the pool. If nil then
	// the job is allocated to a server runner.
	AgentPoolID *resource.TfeID `jsonapi:"attribute" json:"agent_pool_id"`
	// Name of job's organization
	Organization organization.Name `jsonapi:"attribute" json:"organization"`
	// ID of job's workspace
	WorkspaceID resource.TfeID `jsonapi:"attribute" json:"workspace_id"`
	// ID of runner that this job is allocated to. Only set once job enters
	// JobAllocated state.
	RunnerID *resource.TfeID `jsonapi:"attribute" json:"runner_id"`
	// Signaled is non-nil when a cancelation signal has been sent to the job
	// and it is true when it has been forceably canceled.
	Signaled *bool `jsonapi:"attribute" json:"signaled"`
}

func newJob(run *otfrun.Run) *Job {
	return &Job{
		ID:           resource.NewTfeID(resource.JobKind),
		RunID:        run.ID,
		Phase:        run.Phase(),
		Status:       JobUnallocated,
		Organization: run.Organization,
		WorkspaceID:  run.WorkspaceID,
	}
}

func (j *Job) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("job_id", j.ID.String()),
		slog.String("run_id", j.RunID.String()),
		slog.String("workspace_id", j.WorkspaceID.String()),
		slog.Any("organization", j.Organization),
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

func (j *Job) String() string { return j.ID.String() }

func (j *Job) CanAccess(action authz.Action, req *authz.AccessRequest) bool {
	if req == nil {
		// Job cannot carry out site-wide actions
		return false
	}
	if req.Organization != nil && *req.Organization != j.Organization {
		// Job cannot carry out actions on other organizations
		return false
	}
	// Permissible organization actions on same organization
	switch action {
	case authz.GetOrganizationAction, authz.GetEntitlementsAction, authz.GetModuleAction, authz.ListModulesAction:
		return true
	}
	// Permissible workspace actions on same workspace.
	if req.ID != nil && *req.ID == j.WorkspaceID {
		// Allow actions on same workspace as job depending on run phase
		switch action {
		case authz.DownloadStateAction, authz.GetStateVersionAction, authz.GetWorkspaceAction, authz.GetRunAction, authz.ListVariableSetsAction, authz.ListWorkspaceVariablesAction, authz.PutChunkAction, authz.DownloadConfigurationVersionAction, authz.GetPlanFileAction, authz.CancelRunAction:
			// any phase
			return true
		case authz.UploadLockFileAction, authz.UploadPlanFileAction, authz.ApplyRunAction:
			// plan phase
			if j.Phase == internal.PlanPhase {
				return true
			}
		case authz.GetLockFileAction, authz.CreateStateVersionAction:
			// apply phase
			if j.Phase == internal.ApplyPhase {
				return true
			}
		}
		return false
	}
	// If workspace policy is non-nil then that means the job is trying to
	// access *another* workspace. Check the policy to determine if it is
	// allowed to do so.
	if req.WorkspacePolicy != nil {
		switch action {
		case authz.GetStateVersionAction, authz.GetWorkspaceAction, authz.DownloadStateAction:
			if req.WorkspacePolicy.GlobalRemoteState {
				// Job is allowed to retrieve the state of this workspace
				// because the workspace has allowed global remote state
				// sharing.
				return true
			}
		}
	}
	return false
}

func (j *Job) allocate(runnerID resource.TfeID) error {
	if err := j.updateStatus(JobAllocated); err != nil {
		return err
	}
	j.RunnerID = &runnerID
	return nil
}

func (j *Job) reallocate(runnerID resource.TfeID) error {
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
	case runstatus.Planning, runstatus.Applying:
		if run.CancelSignaledAt != nil {
			// run is still in progress but the user has requested it be
			// canceled, so signal job to gracefully cancel current operation
			signal = internal.Bool(false)
		}
	case runstatus.Canceled:
		// run has been canceled so immediately cancel job too
		if err := j.updateStatus(JobCanceled); err != nil {
			return nil, err
		}
	case runstatus.ForceCanceled:
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
