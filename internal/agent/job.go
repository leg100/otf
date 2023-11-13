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

var ErrInvalidJobStateTransition = errors.New("invalid job state transition")

type JobStatus string

const (
	JobUnallocated JobStatus = "unallocated"
	JobAllocated   JobStatus = "allocated"
	JobRunning     JobStatus = "running"
	JobFinished    JobStatus = "finished"
	JobErrored     JobStatus = "errored"
	JobCanceled    JobStatus = "canceled"
)

type signal string

const (
	cancelSignal      signal = "cancel"
	forceCancelSignal signal = "force_cancel"
)

// JobSpec uniquely identifies a job.
type JobSpec struct {
	// ID of the run that this job is for.
	RunID string `json:"run_id"`
	// Phase of run that this job is for.
	Phase internal.PhaseType `json:"phase"`
}

func NewJobSpecFromString(spec string) (JobSpec, error) {
	parts := strings.Split(spec, "-")
	if len(parts) != 2 {
		return JobSpec{}, fmt.Errorf("malformed stringified job spec: %s", spec)
	}
	return JobSpec{RunID: parts[0], Phase: internal.PhaseType(parts[1])}, nil
}

func (j JobSpec) String() string {
	return fmt.Sprintf("%s-%s", j.RunID, j.Phase)
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
	// This indicates whether the run for the job has been signaled, i.e. user
	// has sent a cancelation request.
	signal *signal
	// JWT token. This is only set when a job is newly allocated to an agent.
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
	if j.signal != nil {
		attrs = append(attrs, slog.String("signal", string(*j.signal)))
	}
	return slog.GroupValue(attrs...)
}

func (t *Job) Organizations() []string { return nil }

func (t *Job) IsSiteAdmin() bool   { return true }
func (t *Job) IsOwner(string) bool { return true }

func (t *Job) CanAccessSite(action rbac.Action) bool {
	return false
}

func (t *Job) CanAccessOrganization(action rbac.Action, name string) bool {
	switch action {
	case rbac.GetOrganizationAction, rbac.GetEntitlementsAction, rbac.GetModuleAction, rbac.ListModulesAction:
		return t.Organization == name
	default:
		return false
	}
}

func (t *Job) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// job token is allowed the retrieve the state of the workspace only if:
	// (a) workspace is in the same organization as job token
	// (b) workspace has enabled global remote state (permitting organization-wide
	// state sharing).
	switch action {
	case rbac.GetWorkspaceAction, rbac.GetStateVersionAction, rbac.DownloadStateAction:
		if t.Organization == policy.Organization && policy.GlobalRemoteState {
			return true
		}
	}
	return false
}

func (t *Job) CanAccessTeam(rbac.Action, string) bool {
	// Can't access team level actions
	return false
}

func (j *Job) setSignal(s signal) error {
	if j.Status != JobRunning {
		return errors.New("job can only be signaled when in the JobRunning state")
	}
	j.signal = &s
	return nil
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
