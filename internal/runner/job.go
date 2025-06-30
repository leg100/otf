package runner

import (
	"errors"
	"log/slog"

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
	ID resource.TfeID `jsonapi:"primary,jobs" db:"job_id"`
	// ID of the run that this job is for.
	RunID resource.TfeID `jsonapi:"attribute" json:"run_id" db:"run_id"`
	// Phase of run that this job is for.
	Phase otfrun.PhaseType `jsonapi:"attribute" json:"phase"`
	// Current status of job.
	Status JobStatus `jsonapi:"attribute" json:"status"`
	// ID of agent pool the job's workspace is assigned to use. If non-nil then
	// the job is allocated to an agent runner belonging to the pool. If nil then
	// the job is allocated to a server runner.
	AgentPoolID *resource.TfeID `jsonapi:"attribute" json:"agent_pool_id" db:"agent_pool_id"`
	// Name of job's organization
	Organization organization.Name `jsonapi:"attribute" json:"organization" db:"organization_name"`
	// ID of job's workspace
	WorkspaceID resource.TfeID `jsonapi:"attribute" json:"workspace_id" db:"workspace_id"`
	// ID of runner that this job is allocated to. Only set once job enters
	// JobAllocated state.
	RunnerID *resource.TfeID `jsonapi:"attribute" json:"runner_id" db:"runner_id"`
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
	return slog.GroupValue(attrs...)
}

func (j *Job) String() string { return j.ID.String() }

func (j *Job) CanAccess(action authz.Action, req authz.Request) bool {
	if req.Kind() == resource.SiteKind {
		// Job cannot carry out site-wide actions
		return false
	}
	if req.Organization() != nil && req.Organization().String() != j.Organization.String() {
		// Job cannot carry out actions on other organizations
		return false
	}
	// Permissible organization actions on same organization
	switch action {
	case authz.GetOrganizationAction, authz.GetEntitlementsAction, authz.GetModuleAction, authz.ListModulesAction:
		return true
	}
	// Permissible workspace actions on same workspace.
	if req.Workspace() == j.WorkspaceID {
		// Allow actions on same workspace as job depending on run phase
		switch action {
		case authz.DownloadStateAction, authz.GetStateVersionAction, authz.GetWorkspaceAction, authz.GetRunAction, authz.ListVariableSetsAction, authz.ListWorkspaceVariablesAction, authz.PutChunkAction, authz.DownloadConfigurationVersionAction, authz.GetPlanFileAction, authz.CancelRunAction:
			// any phase
			return true
		case authz.UploadLockFileAction, authz.UploadPlanFileAction, authz.ApplyRunAction:
			// plan phase
			if j.Phase == otfrun.PlanPhase {
				return true
			}
		case authz.GetLockFileAction, authz.CreateStateVersionAction:
			// apply phase
			if j.Phase == otfrun.ApplyPhase {
				return true
			}
		}
		return false
	}
	// Check workspace policy if there is one - if the job is attempting to
	// access the state of another workspace then the policy determines whether
	// it's allowed to do so.
	if req.WorkspacePolicy != nil {
		return req.WorkspacePolicy.Check(j.ID, action)
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
func (j *Job) cancel(run *otfrun.Run) (signal bool, force bool, err error) {
	switch run.Status {
	case runstatus.Planning, runstatus.Applying:
		if run.CancelSignaledAt != nil {
			// run is still in progress but the user has requested it be
			// canceled, so signal job to gracefully cancel current operation
			return true, false, nil
		}
	case runstatus.Canceled:
		// run has been canceled so immediately cancel job too but don't send a
		// signal.
		if err := j.updateStatus(JobCanceled); err != nil {
			return false, false, err
		}
	case runstatus.ForceCanceled:
		// run has been forceably canceled, so both signal job to forcefully
		// cancel current operation, and immediately cancel job.
		if err := j.updateStatus(JobCanceled); err != nil {
			return false, false, err
		}
		return true, true, nil
	}
	return false, false, nil
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
