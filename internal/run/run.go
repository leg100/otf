// Package run is responsible for OTF runs, the primary mechanism for executing
// terraform
package run

import (
	"errors"
	"fmt"
	"time"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/workspace"
)

const (
	PlanFormatBinary = "bin"  // plan file in binary format
	PlanFormatJSON   = "json" // plan file in json format

	// defaultRefresh specifies that the state be refreshed prior to running a
	// plan
	defaultRefresh = true
)

var ErrInvalidRunStateTransition = errors.New("invalid run state transition")

type (
	PlanFormat string

	// Run is a terraform run.
	Run struct {
		ID                     string
		CreatedAt              time.Time
		IsDestroy              bool
		ForceCancelAvailableAt *time.Time
		Message                string
		Organization           string
		Refresh                bool
		RefreshOnly            bool
		ReplaceAddrs           []string
		PositionInQueue        int
		TargetAddrs            []string
		AutoApply              bool
		Speculative            bool
		Status                 internal.RunStatus
		StatusTimestamps       []RunStatusTimestamp
		WorkspaceID            string
		ConfigurationVersionID string
		ExecutionMode          workspace.ExecutionMode
		Plan                   Phase
		Apply                  Phase

		Latest bool    // is latest run for workspace
		Commit *string // commit sha that triggered this run
	}

	// RunList represents a list of runs.
	RunList struct {
		*internal.Pagination
		Items []*Run
	}

	RunStatusTimestamp struct {
		Status    internal.RunStatus
		Timestamp time.Time
	}

	// RunCreateOptions represents the options for creating a new run. See
	// dto.RunCreateOptions for further detail.
	RunCreateOptions struct {
		IsDestroy              *bool
		Refresh                *bool
		RefreshOnly            *bool
		Message                *string
		ConfigurationVersionID *string
		TargetAddrs            []string
		ReplaceAddrs           []string
		AutoApply              *bool
	}

	// RunListOptions are options for paginating and filtering a list of runs
	RunListOptions struct {
		internal.ListOptions
		// Filter by run statuses (with an implicit OR condition)
		Statuses []internal.RunStatus `schema:"statuses,omitempty"`
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

	// WatchOptions filters events returned by the Watch endpoint.
	WatchOptions struct {
		Organization *string `schema:"organization_name,omitempty"` // filter by organization name
		WorkspaceID  *string `schema:"workspace_id,omitempty"`      // filter by workspace ID; mutually exclusive with organization filter
	}
)

// NewRun creates a new run with defaults.
func NewRun(cv *configversion.ConfigurationVersion, ws *workspace.Workspace, opts RunCreateOptions) *Run {
	run := Run{
		ID:                     internal.NewID("run"),
		CreatedAt:              internal.CurrentTimestamp(),
		Refresh:                defaultRefresh,
		Organization:           ws.Organization,
		ConfigurationVersionID: cv.ID,
		WorkspaceID:            ws.ID,
		Speculative:            cv.Speculative,
		ReplaceAddrs:           opts.ReplaceAddrs,
		TargetAddrs:            opts.TargetAddrs,
		ExecutionMode:          ws.ExecutionMode,
		AutoApply:              ws.AutoApply,
	}
	run.Plan = NewPhase(run.ID, internal.PlanPhase)
	run.Apply = NewPhase(run.ID, internal.ApplyPhase)
	run.updateStatus(internal.RunPending)

	if opts.IsDestroy != nil {
		run.IsDestroy = *opts.IsDestroy
	}
	if opts.Message != nil {
		run.Message = *opts.Message
	}
	if opts.Refresh != nil {
		run.Refresh = *opts.Refresh
	}
	if opts.AutoApply != nil {
		run.AutoApply = *opts.AutoApply
	}
	if cv.IngressAttributes != nil {
		run.Commit = &cv.IngressAttributes.CommitSHA
	}
	return &run
}

func (r *Run) Queued() bool {
	return r.Status == internal.RunPlanQueued || r.Status == internal.RunApplyQueued
}

func (r *Run) HasChanges() bool {
	return r.Plan.HasChanges()
}

func (r *Run) PlanOnly() bool {
	return r.Status == internal.RunPlannedAndFinished
}

// HasApply determines whether the run has started applying yet.
func (r *Run) HasApply() bool {
	_, err := r.Apply.StatusTimestamp(PhaseRunning)
	return err == nil
}

// Phase returns the current phase.
func (r *Run) Phase() internal.PhaseType {
	switch r.Status {
	case internal.RunPending:
		return internal.PendingPhase
	case internal.RunPlanQueued, internal.RunPlanning, internal.RunPlanned:
		return internal.PlanPhase
	case internal.RunApplyQueued, internal.RunApplying, internal.RunApplied:
		return internal.ApplyPhase
	default:
		return internal.UnknownPhase
	}
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.Discardable() {
		return internal.ErrRunDiscardNotAllowed
	}
	r.updateStatus(internal.RunDiscarded)

	if r.Status == internal.RunPending {
		r.Plan.UpdateStatus(PhaseUnreachable)
	}
	r.Apply.UpdateStatus(PhaseUnreachable)

	return nil
}

// Cancel run. Returns a boolean indicating whether a cancel request should be
// enqueued (for an agent to kill an in progress process)
func (r *Run) Cancel() (enqueue bool, err error) {
	if !r.Cancelable() {
		return false, internal.ErrRunCancelNotAllowed
	}
	// permit run to be force canceled after a cool off period of 10 seconds has
	// elapsed.
	tenSecondsFromNow := internal.CurrentTimestamp().Add(10 * time.Second)
	r.ForceCancelAvailableAt = &tenSecondsFromNow

	switch r.Status {
	case internal.RunPending:
		r.Plan.UpdateStatus(PhaseUnreachable)
		r.Apply.UpdateStatus(PhaseUnreachable)
	case internal.RunPlanQueued, internal.RunPlanning:
		r.Plan.UpdateStatus(PhaseCanceled)
		r.Apply.UpdateStatus(PhaseUnreachable)
	case internal.RunApplyQueued, internal.RunApplying:
		r.Apply.UpdateStatus(PhaseCanceled)
	}

	if r.Status == internal.RunPlanning || r.Status == internal.RunApplying {
		enqueue = true
	}

	r.updateStatus(internal.RunCanceled)

	return enqueue, nil
}

// ForceCancel force cancels a run. A cool-off period of 10 seconds must have
// elapsed following a cancelation request before a run can be force canceled.
func (r *Run) ForceCancel() error {
	if r.ForceCancelAvailableAt != nil && time.Now().After(*r.ForceCancelAvailableAt) {
		r.updateStatus(internal.RunCanceled)
		return nil
	}
	return internal.ErrRunForceCancelNotAllowed
}

// Done determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (r *Run) Done() bool {
	switch r.Status {
	case internal.RunApplied, internal.RunPlannedAndFinished, internal.RunDiscarded, internal.RunCanceled, internal.RunErrored:
		return true
	default:
		return false
	}
}

// EnqueuePlan enqueues a plan for the run. It also sets the run as the latest
// run for its workspace (speculative runs are ignored).
func (r *Run) EnqueuePlan() error {
	if r.Status != internal.RunPending {
		return fmt.Errorf("cannot enqueue run with status %s", r.Status)
	}
	r.updateStatus(internal.RunPlanQueued)
	r.Plan.UpdateStatus(PhaseQueued)

	return nil
}

func (*Run) CanAccessSite(action rbac.Action) bool {
	// run cannot carry out site-level actions
	return false
}

func (r *Run) CanAccessOrganization(action rbac.Action, name string) bool {
	// run cannot access organization-level resources
	return false
}

func (r *Run) CanAccessWorkspace(action rbac.Action, policy *internal.WorkspacePolicy) bool {
	// run can access anything within its workspace
	return r.WorkspaceID == policy.WorkspaceID
}

func (r *Run) EnqueueApply() error {
	if r.Status != internal.RunPlanned {
		return fmt.Errorf("cannot apply run")
	}
	r.updateStatus(internal.RunApplyQueued)
	r.Apply.UpdateStatus(PhaseQueued)
	return nil
}

func (r *Run) StatusTimestamp(status internal.RunStatus) (time.Time, error) {
	for _, rst := range r.StatusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, internal.ErrStatusTimestampNotFound
}

// Start a run phase
func (r *Run) Start(phase internal.PhaseType) error {
	switch r.Status {
	case internal.RunPlanQueued:
		r.updateStatus(internal.RunPlanning)
		r.Plan.UpdateStatus(PhaseRunning)
	case internal.RunApplyQueued:
		r.updateStatus(internal.RunApplying)
		r.Apply.UpdateStatus(PhaseRunning)
	case internal.RunPlanning, internal.RunApplying:
		return ErrPhaseAlreadyStarted
	default:
		return ErrInvalidRunStateTransition
	}
	return nil
}

// Finish updates the run to reflect its plan or apply phase having finished.
func (r *Run) Finish(phase internal.PhaseType, opts PhaseFinishOptions) error {
	if r.Status == internal.RunCanceled {
		// run was canceled before the phase finished so nothing more to do.
		return nil
	}
	switch phase {
	case internal.PlanPhase:
		if r.Status != internal.RunPlanning {
			return ErrInvalidRunStateTransition
		}
		if opts.Errored {
			r.updateStatus(internal.RunErrored)
			r.Plan.UpdateStatus(PhaseErrored)
			r.Apply.UpdateStatus(PhaseUnreachable)
			return nil
		}

		r.updateStatus(internal.RunPlanned)
		r.Plan.UpdateStatus(PhaseFinished)

		if !r.HasChanges() || r.Speculative {
			r.updateStatus(internal.RunPlannedAndFinished)
			r.Apply.UpdateStatus(PhaseUnreachable)
		} else if r.AutoApply {
			return r.EnqueueApply()
		}
		return nil
	case internal.ApplyPhase:
		if r.Status != internal.RunApplying {
			return ErrInvalidRunStateTransition
		}
		if opts.Errored {
			r.updateStatus(internal.RunErrored)
			r.Apply.UpdateStatus(PhaseErrored)
		} else {
			r.updateStatus(internal.RunApplied)
			r.Apply.UpdateStatus(PhaseFinished)
		}
		return nil
	default:
		return fmt.Errorf("unknown phase")
	}
}

func (r *Run) updateStatus(status internal.RunStatus) {
	r.Status = status
	r.StatusTimestamps = append(r.StatusTimestamps, RunStatusTimestamp{
		Status:    status,
		Timestamp: internal.CurrentTimestamp(),
	})
}

// Discardable determines whether run can be discarded.
func (r *Run) Discardable() bool {
	switch r.Status {
	case internal.RunPending, internal.RunPlanned:
		return true
	default:
		return false
	}
}

// Cancelable determines whether run can be cancelled.
func (r *Run) Cancelable() bool {
	switch r.Status {
	case internal.RunPending, internal.RunPlanQueued, internal.RunPlanning, internal.RunApplyQueued, internal.RunApplying:
		return true
	default:
		return false
	}
}

// Confirmable determines whether run can be confirmed.
func (r *Run) Confirmable() bool {
	switch r.Status {
	case internal.RunPlanned:
		return true
	default:
		return false
	}
}
