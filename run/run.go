package run

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

const (
	// DefaultRefresh specifies that the state be refreshed prior to running a
	// plan
	DefaultRefresh = true
)

var (
	ErrRunDiscardNotAllowed      = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed       = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed  = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")
	ErrInvalidRunGetOptions      = errors.New("invalid run get options")
	ErrInvalidRunStateTransition = errors.New("invalid run state transition")
)

type Run struct {
	id                     string
	createdAt              time.Time
	forceCancelAvailableAt *time.Time
	isDestroy              bool
	message                string
	executionMode          otf.ExecutionMode
	positionInQueue        int
	refresh                bool
	refreshOnly            bool
	autoApply              bool
	speculative            bool
	status                 otf.RunStatus
	statusTimestamps       []RunStatusTimestamp
	replaceAddrs           []string
	targetAddrs            []string
	organization           string
	workspaceID            string
	configurationVersionID string
	latest                 bool    // is latest run for workspace
	commit                 *string // commit sha that triggered this run
	plan                   *Plan
	apply                  *Apply
}

func (r *Run) ID() string                             { return r.id }
func (r *Run) CreatedAt() time.Time                   { return r.createdAt }
func (r *Run) String() string                         { return r.id }
func (r *Run) IsDestroy() bool                        { return r.isDestroy }
func (r *Run) ForceCancelAvailableAt() *time.Time     { return r.forceCancelAvailableAt }
func (r *Run) Message() string                        { return r.message }
func (r *Run) Organization() string                   { return r.organization }
func (r *Run) Refresh() bool                          { return r.refresh }
func (r *Run) RefreshOnly() bool                      { return r.refreshOnly }
func (r *Run) ReplaceAddrs() []string                 { return r.replaceAddrs }
func (r *Run) TargetAddrs() []string                  { return r.targetAddrs }
func (r *Run) AutoApply() bool                        { return r.autoApply }
func (r *Run) Speculative() bool                      { return r.speculative }
func (r *Run) Status() otf.RunStatus                  { return r.status }
func (r *Run) StatusTimestamps() []RunStatusTimestamp { return r.statusTimestamps }
func (r *Run) WorkspaceID() string                    { return r.workspaceID }
func (r *Run) ConfigurationVersionID() string         { return r.configurationVersionID }
func (r *Run) Plan() *Plan                            { return r.plan }
func (r *Run) Apply() *Apply                          { return r.apply }
func (r *Run) ExecutionMode() otf.ExecutionMode       { return r.executionMode }
func (r *Run) Commit() *string                        { return r.commit }

// TODO: is this necessary?
func (r *Run) RunID() string { return r.id }

// Latest determines whether run is the latest run for a workspace, i.e.
// its current run, or the most recent current run.
func (r *Run) Latest() bool { return r.latest }

func (r *Run) Queued() bool {
	return r.status == otf.RunPlanQueued || r.status == otf.RunApplyQueued
}

func (r *Run) HasChanges() bool {
	return r.plan.HasChanges()
}

func (r *Run) PlanOnly() bool {
	return r.status == otf.RunPlannedAndFinished
}

// HasApply determines whether the run has started applying yet.
func (r *Run) HasApply() bool {
	_, err := r.Apply().StatusTimestamp(otf.PhaseRunning)
	return err == nil
}

// Phase returns the current phase.
func (r *Run) Phase() otf.PhaseType {
	switch r.status {
	case otf.RunPending:
		return otf.PendingPhase
	case otf.RunPlanQueued, otf.RunPlanning, otf.RunPlanned:
		return otf.PlanPhase
	case otf.RunApplyQueued, otf.RunApplying, otf.RunApplied:
		return otf.ApplyPhase
	default:
		return otf.UnknownPhase
	}
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.Discardable() {
		return ErrRunDiscardNotAllowed
	}
	r.updateStatus(otf.RunDiscarded)

	if r.status == otf.RunPending {
		r.plan.updateStatus(otf.PhaseUnreachable)
	}
	r.apply.updateStatus(otf.PhaseUnreachable)

	return nil
}

// Cancel run. Returns a boolean indicating whether a cancel request should be
// enqueued (for an agent to kill an in progress process)
func (r *Run) Cancel() (enqueue bool, err error) {
	if !r.Cancelable() {
		return false, ErrRunCancelNotAllowed
	}
	// permit run to be force canceled after a cool off period of 10 seconds has
	// elapsed.
	tenSecondsFromNow := otf.CurrentTimestamp().Add(10 * time.Second)
	r.forceCancelAvailableAt = &tenSecondsFromNow

	switch r.status {
	case otf.RunPending:
		r.plan.updateStatus(otf.PhaseUnreachable)
		r.apply.updateStatus(otf.PhaseUnreachable)
	case otf.RunPlanQueued, otf.RunPlanning:
		r.plan.updateStatus(otf.PhaseCanceled)
		r.apply.updateStatus(otf.PhaseUnreachable)
	case otf.RunApplyQueued, otf.RunApplying:
		r.apply.updateStatus(otf.PhaseCanceled)
	}

	if r.status == otf.RunPlanning || r.status == otf.RunApplying {
		enqueue = true
	}

	r.updateStatus(otf.RunCanceled)

	return enqueue, nil
}

// ForceCancel force cancels a run. A cool-off period of 10 seconds must have
// elapsed following a cancelation request before a run can be force canceled.
func (r *Run) ForceCancel() error {
	if r.forceCancelAvailableAt != nil && time.Now().After(*r.forceCancelAvailableAt) {
		r.updateStatus(otf.RunCanceled)
		return nil
	}
	return ErrRunForceCancelNotAllowed
}

// Done determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (r *Run) Done() bool {
	switch r.Status() {
	case otf.RunApplied, otf.RunPlannedAndFinished, otf.RunDiscarded, otf.RunCanceled, otf.RunErrored:
		return true
	default:
		return false
	}
}

// EnqueuePlan enqueues a plan for the run. It also sets the run as the latest
// run for its workspace (speculative runs are ignored).
func (r *Run) EnqueuePlan() error {
	if r.status != otf.RunPending {
		return fmt.Errorf("cannot enqueue run with status %s", r.status)
	}
	r.updateStatus(otf.RunPlanQueued)
	r.plan.updateStatus(otf.PhaseQueued)

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

func (r *Run) CanAccessWorkspace(action rbac.Action, policy *otf.WorkspacePolicy) bool {
	// run can access anything within its workspace
	return r.workspaceID == policy.WorkspaceID
}

func (r *Run) EnqueueApply() error {
	if r.status != otf.RunPlanned {
		return fmt.Errorf("cannot apply run")
	}
	r.updateStatus(otf.RunApplyQueued)
	r.apply.updateStatus(otf.PhaseQueued)
	return nil
}

func (r *Run) StatusTimestamp(status otf.RunStatus) (time.Time, error) {
	for _, rst := range r.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, otf.ErrStatusTimestampNotFound
}

// Start a run phase
func (r *Run) Start(phase otf.PhaseType) error {
	switch r.status {
	case otf.RunPlanQueued:
		return r.startPlan()
	case otf.RunApplyQueued:
		return r.startApply()
	case otf.RunPlanning, otf.RunApplying:
		return otf.ErrPhaseAlreadyStarted
	default:
		return ErrInvalidRunStateTransition
	}
}

// Finish updates the run to reflect its plan or apply phase having finished.
func (r *Run) Finish(phase otf.PhaseType, opts otf.PhaseFinishOptions) error {
	if r.status == otf.RunCanceled {
		// run was canceled before the phase finished so nothing more to do.
		return nil
	}
	switch phase {
	case otf.PlanPhase:
		return r.finishPlan(opts)
	case otf.ApplyPhase:
		return r.finishApply(opts)
	default:
		return fmt.Errorf("unknown phase")
	}
}

func (r *Run) toValue() otf.Run {
	return otf.Run{ID: r.id}
}

func (r *Run) startPlan() error {
	if r.status != otf.RunPlanQueued {
		return ErrInvalidRunStateTransition
	}
	r.updateStatus(otf.RunPlanning)
	r.plan.updateStatus(otf.PhaseRunning)
	return nil
}

func (r *Run) startApply() error {
	if r.status != otf.RunApplyQueued {
		return ErrInvalidRunStateTransition
	}
	r.updateStatus(otf.RunApplying)
	r.apply.updateStatus(otf.PhaseRunning)
	return nil
}

func (r *Run) finishPlan(opts otf.PhaseFinishOptions) error {
	if r.status != otf.RunPlanning {
		return ErrInvalidRunStateTransition
	}
	if opts.Errored {
		r.updateStatus(otf.RunErrored)
		r.plan.updateStatus(otf.PhaseErrored)
		r.apply.updateStatus(otf.PhaseUnreachable)
		return nil
	}

	r.updateStatus(otf.RunPlanned)
	r.plan.updateStatus(otf.PhaseFinished)

	if !r.HasChanges() || r.Speculative() {
		r.updateStatus(otf.RunPlannedAndFinished)
		r.apply.updateStatus(otf.PhaseUnreachable)
	} else if r.autoApply {
		return r.EnqueueApply()
	}
	return nil
}

func (r *Run) finishApply(opts otf.PhaseFinishOptions) error {
	if r.status != otf.RunApplying {
		return ErrInvalidRunStateTransition
	}
	if opts.Errored {
		r.updateStatus(otf.RunErrored)
		r.apply.updateStatus(otf.PhaseErrored)
	} else {
		r.updateStatus(otf.RunApplied)
		r.apply.updateStatus(otf.PhaseFinished)
	}
	return nil
}

func (r *Run) updateStatus(status otf.RunStatus) {
	r.status = status
	r.statusTimestamps = append(r.statusTimestamps, RunStatusTimestamp{
		Status:    status,
		Timestamp: otf.CurrentTimestamp(),
	})
}

// Discardable determines whether run can be discarded.
func (r *Run) Discardable() bool {
	switch r.Status() {
	case otf.RunPending, otf.RunPlanned:
		return true
	default:
		return false
	}
}

// Cancelable determines whether run can be cancelled.
func (r *Run) Cancelable() bool {
	switch r.Status() {
	case otf.RunPending, otf.RunPlanQueued, otf.RunPlanning, otf.RunPlanned, otf.RunApplyQueued, otf.RunApplying:
		return true
	default:
		return false
	}
}

// Confirmable determines whether run can be confirmed.
func (r *Run) Confirmable() bool {
	switch r.Status() {
	case otf.RunPlanned:
		return true
	default:
		return false
	}
}

type RunStatusTimestamp struct {
	Status    otf.RunStatus
	Timestamp time.Time
}

// RunCreateOptions represents the options for creating a new run. See
// dto.RunCreateOptions for further detail.
type RunCreateOptions struct {
	IsDestroy              *bool
	Refresh                *bool
	RefreshOnly            *bool
	Message                *string
	ConfigurationVersionID *string
	TargetAddrs            []string
	ReplaceAddrs           []string
	AutoApply              *bool
}

// RunList represents a list of runs.
type RunList struct {
	*otf.Pagination
	Items []*Run
}

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	otf.ListOptions
	// Filter by run statuses (with an implicit OR condition)
	Statuses []otf.RunStatus `schema:"statuses,omitempty"`
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
