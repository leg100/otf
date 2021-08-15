package ots

import (
	"errors"
	"time"

	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

const (
	// DefaultRefresh specifies that the state be refreshed prior to running a
	// plan
	DefaultRefresh = true
)

var (
	ErrRunDiscardNotAllowed     = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed      = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")

	ErrInvalidRunGetOptions = errors.New("invalid run get options")

	// ActiveRunStatuses are those run statuses that deem a run to be active.
	// There can only be at most one active run for a workspace.
	ActiveRunStatuses = []tfe.RunStatus{
		tfe.RunApplyQueued,
		tfe.RunApplying,
		tfe.RunConfirmed,
		tfe.RunPlanQueued,
		tfe.RunPlanned,
		tfe.RunPlanning,
	}
)

type Run struct {
	ID string

	gorm.Model

	ForceCancelAvailableAt time.Time
	IsDestroy              bool
	Message                string
	Permissions            *tfe.RunPermissions
	PositionInQueue        int
	Refresh                bool
	RefreshOnly            bool
	Status                 tfe.RunStatus
	StatusTimestamps       *tfe.RunStatusTimestamps
	ReplaceAddrs           []string
	TargetAddrs            []string

	// Relations
	Plan                 *Plan
	Apply                *Apply
	Workspace            *Workspace
	ConfigurationVersion *ConfigurationVersion
}

// RunFactory is a factory for constructing Run objects.
type RunFactory struct {
	ConfigurationVersionService ConfigurationVersionService
	WorkspaceService            WorkspaceService
}

// RunService implementations allow interactions with runs
type RunService interface {
	Create(opts *tfe.RunCreateOptions) (*Run, error)
	Get(id string) (*Run, error)
	List(opts RunListOptions) (*RunList, error)
	GetQueued(opts tfe.RunListOptions) (*RunList, error)
	Apply(id string, opts *tfe.RunApplyOptions) error
	Discard(id string, opts *tfe.RunDiscardOptions) error
	Cancel(id string, opts *tfe.RunCancelOptions) error
	ForceCancel(id string, opts *tfe.RunForceCancelOptions) error
	UpdateStatus(id string, status tfe.RunStatus) (*Run, error)
	UpdatePlanStatus(id string, status tfe.PlanStatus) (*Run, error)
	UpdateApplyStatus(id string, status tfe.ApplyStatus) (*Run, error)
	GetPlanLogs(id string, opts PlanLogOptions) ([]byte, error)
	UploadPlanLogs(id string, logs []byte) error
	GetApplyLogs(id string, opts ApplyLogOptions) ([]byte, error)
	UploadApplyLogs(id string, logs []byte) error
	FinishPlan(id string, opts PlanFinishOptions) (*Run, error)
	FinishApply(id string, opts ApplyFinishOptions) (*Run, error)
	GetPlanJSON(id string) ([]byte, error)
	GetPlanFile(id string) ([]byte, error)
}

// RunStore implementations persist Run objects.
type RunStore interface {
	Create(run *Run) (*Run, error)
	Get(opts RunGetOptions) (*Run, error)
	List(opts RunListOptions) (*RunList, error)
	// TODO: add support for a special error type that tells update to skip
	// updates - useful when fn checks current fields and decides not to update
	Update(id string, fn func(*Run) error) (*Run, error)
}

// RunList represents a list of runs.
type RunList struct {
	*tfe.Pagination
	Items []*Run
}

// RunGetOptions are options for retrieving a single Run. Either ID *or* ApplyID
// must be specfiied.
type RunGetOptions struct {
	// ID of run to retrieve
	ID *string

	// Get run via apply ID
	ApplyID *string

	// Get run via plan ID
	PlanID *string
}

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	tfe.RunListOptions

	// Filter by run statuses (with an implicit OR condition)
	Statuses []tfe.RunStatus

	// Filter by workspace ID
	WorkspaceID *string
}

// FinishPlan updates the state of a run to reflect its plan having finished
func (r *Run) FinishPlan(opts PlanFinishOptions) error {
	r.Plan.ResourceAdditions = opts.ResourceAdditions
	r.Plan.ResourceChanges = opts.ResourceChanges
	r.Plan.ResourceDestructions = opts.ResourceDestructions

	r.UpdatePlanStatus(tfe.PlanFinished)

	return nil
}

// FinishApply updates the state of a run to reflect its plan having finished
func (r *Run) FinishApply(opts ApplyFinishOptions) error {
	r.Apply.ResourceAdditions = opts.ResourceAdditions
	r.Apply.ResourceChanges = opts.ResourceChanges
	r.Apply.ResourceDestructions = opts.ResourceDestructions

	r.UpdateApplyStatus(tfe.ApplyFinished)

	return nil
}

// UpdateStatusToPlanQueued updates a run's status to indicate its plan has been
// queued.
func (r *Run) UpdateStatusToPlanQueued() {
	r.Status = tfe.RunPlanQueued
	r.StatusTimestamps.PlanQueuedAt = TimeNow()
	r.Plan.Status = tfe.PlanQueued
	r.Plan.StatusTimestamps.QueuedAt = TimeNow()
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.IsDiscardable() {
		return ErrRunDiscardNotAllowed
	}

	r.Status = tfe.RunDiscarded
	r.StatusTimestamps.DiscardedAt = TimeNow()

	// TODO: update plan status

	r.Apply.Status = tfe.ApplyUnreachable
	r.Apply.StatusTimestamps.CanceledAt = TimeNow()

	return nil
}

// IssueCancel updates the state of a run to reflect a cancel request having
// been issued.
func (r *Run) IssueCancel() error {
	if !r.IsCancelable() {
		return ErrRunCancelNotAllowed
	}

	// Run can be forcefully cancelled after a cool-off period of ten seconds
	r.ForceCancelAvailableAt = time.Now().Add(10 * time.Second)

	return nil
}

// ForceCancel updates the state of a run to reflect it having been forcefully
// cancelled.
func (r *Run) ForceCancel() error {
	if !r.IsForceCancelable() {
		return ErrRunForceCancelNotAllowed
	}

	// Put plan or apply into cancel state
	switch r.Status {
	case tfe.RunPlanQueued, tfe.RunPlanning:
		r.Plan.Status = tfe.PlanCanceled
		r.Plan.StatusTimestamps.CanceledAt = TimeNow()
	case tfe.RunApplyQueued, tfe.RunApplying:
		r.Apply.Status = tfe.ApplyCanceled
		r.Apply.StatusTimestamps.CanceledAt = TimeNow()
	}

	// Put run into a cancel state
	r.Status = tfe.RunCanceled
	r.StatusTimestamps.CanceledAt = TimeNow()

	return nil
}

// Actions lists which actions are currently invokable.
func (r *Run) Actions() *tfe.RunActions {
	return &tfe.RunActions{
		IsCancelable:      r.IsCancelable(),
		IsConfirmable:     r.IsConfirmable(),
		IsForceCancelable: r.IsForceCancelable(),
		IsDiscardable:     r.IsDiscardable(),
	}
}

// IsCancelable determines whether run can be cancelled.
func (r *Run) IsCancelable() bool {
	switch r.Status {
	case tfe.RunPending, tfe.RunPlanQueued, tfe.RunPlanning, tfe.RunApplyQueued, tfe.RunApplying:
		return true
	default:
		return false
	}
}

// IsConfirmable determines whether run can be confirmed.
func (r *Run) IsConfirmable() bool {
	switch r.Status {
	case tfe.RunPlanned:
		return true
	default:
		return false
	}
}

// IsDiscardable determines whether run can be discarded.
func (r *Run) IsDiscardable() bool {
	switch r.Status {
	case tfe.RunPending, tfe.RunPolicyChecked, tfe.RunPolicyOverride, tfe.RunPlanned:
		return true
	default:
		return false
	}
}

// IsForceCancelable determines whether a run can be forcibly cancelled.
func (r *Run) IsForceCancelable() bool {
	return r.IsCancelable() && !r.ForceCancelAvailableAt.IsZero() && time.Now().After(r.ForceCancelAvailableAt)
}

// IsActive determines whether run is currently the active run on a workspace,
// i.e. it is neither finished nor pending
func (r *Run) IsActive() bool {
	if r.IsDone() || r.Status == tfe.RunPending {
		return false
	}
	return true
}

// IsDone determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (r *Run) IsDone() bool {
	switch r.Status {
	case tfe.RunApplied, tfe.RunPlannedAndFinished, tfe.RunDiscarded, tfe.RunCanceled, tfe.RunErrored:
		return true
	default:
		return false
	}
}

type ErrInvalidRunStatusTransition struct {
	From tfe.RunStatus
	To   tfe.RunStatus
}

func (e ErrInvalidRunStatusTransition) Error() string {
	return fmt.Sprintf("invalid run status transition from %s to %s", e.From, e.To)
}

// UpdateStatus updates the status of the run.
func (r *Run) UpdateStatus(status tfe.RunStatus) error {
	switch status {
	case tfe.RunPlanQueued:
		if r.Status != tfe.RunPending {
			return ErrInvalidRunStatusTransition{From: r.Status, To: status}
		}
		r.UpdateStatusToPlanQueued()
	case tfe.RunDiscarded:
		return r.Discard()
	case tfe.RunApplyQueued:
		r.Status = tfe.RunApplyQueued
		r.StatusTimestamps.ApplyQueuedAt = TimeNow()

		r.Apply.Status = tfe.ApplyQueued
		r.Apply.StatusTimestamps.QueuedAt = TimeNow()
	case tfe.RunApplied:
		r.Status = tfe.RunApplied
		r.StatusTimestamps.AppliedAt = TimeNow()

		r.Apply.Status = tfe.ApplyFinished
		r.Apply.StatusTimestamps.FinishedAt = TimeNow()
	case tfe.RunErrored:
		r.Status = tfe.RunErrored
		r.StatusTimestamps.ErroredAt = TimeNow()

		// TODO: determine whether to set status to errored on plan or apply
	default:
		// TODO: log no matching status

		// Don't set a status or timestamp
		return nil
	}

	return nil
}

func (r *Run) IsSpeculative() bool {
	return r.ConfigurationVersion.Speculative
}

// UpdatePlanStatus updates the status of the plan. It'll also update the
// appropriate timestamp and set any other appropriate fields for the given
// status.
func (r *Run) UpdatePlanStatus(status tfe.PlanStatus) {
	// Copy timestamps from plan
	timestamps := &tfe.PlanStatusTimestamps{}
	if r.StatusTimestamps != nil {
		timestamps = r.Plan.StatusTimestamps
	}

	switch status {
	case tfe.PlanQueued:
		timestamps.QueuedAt = TimeNow()
		r.UpdateStatus(tfe.RunPlanQueued)
	case tfe.PlanRunning:
		timestamps.StartedAt = TimeNow()
		r.UpdateStatus(tfe.RunPlanning)
	case tfe.PlanCanceled:
		timestamps.CanceledAt = TimeNow()
	case tfe.PlanErrored:
		r.UpdateStatus(tfe.RunErrored)
		timestamps.ErroredAt = TimeNow()
	case tfe.PlanFinished:

		if r.ConfigurationVersion.Speculative {
			r.Status = tfe.RunPlannedAndFinished
			r.StatusTimestamps.PlannedAndFinishedAt = TimeNow()
			timestamps.FinishedAt = TimeNow()
		} else {
			r.Status = tfe.RunPlanned
			timestamps.FinishedAt = TimeNow()
		}
	default:
		// Don't set a timestamp
		return
	}

	r.Plan.Status = status

	// Set timestamps on plan
	r.Plan.StatusTimestamps = timestamps
}

// UpdateApplyStatus updates the status of the apply. It'll also update the
// appropriate timestamp and set any other appropriate fields for the given
// status.
func (r *Run) UpdateApplyStatus(status tfe.ApplyStatus) {
	// Copy timestamps from apply
	timestamps := &tfe.ApplyStatusTimestamps{}
	if r.StatusTimestamps != nil {
		timestamps = r.Apply.StatusTimestamps
	}

	switch status {
	case tfe.ApplyFinished:
		timestamps.FinishedAt = TimeNow()
		r.UpdateStatus(tfe.RunApplied)
	case tfe.ApplyRunning:
		timestamps.StartedAt = TimeNow()
		r.UpdateStatus(tfe.RunApplying)
	case tfe.ApplyQueued:
		timestamps.QueuedAt = TimeNow()
		r.UpdateStatus(tfe.RunApplyQueued)
	case tfe.ApplyCanceled:
		timestamps.CanceledAt = TimeNow()
	case tfe.ApplyErrored:
		timestamps.ErroredAt = TimeNow()
		r.UpdateStatus(tfe.RunErrored)
	default:
		// Don't set a timestamp
		return
	}

	r.Apply.Status = status

	// Set timestamps on apply
	r.Apply.StatusTimestamps = timestamps
}

// NewRun constructs a run object.
func (f *RunFactory) NewRun(opts *tfe.RunCreateOptions) (*Run, error) {
	if opts.Workspace == nil {
		return nil, errors.New("workspace is required")
	}

	run := Run{
		ID: GenerateID("run"),
		Permissions: &tfe.RunPermissions{
			CanForceCancel:  true,
			CanApply:        true,
			CanCancel:       true,
			CanDiscard:      true,
			CanForceExecute: true,
		},
		Refresh:      DefaultRefresh,
		ReplaceAddrs: opts.ReplaceAddrs,
		TargetAddrs:  opts.TargetAddrs,
		Status:       tfe.RunPending,
		StatusTimestamps: &tfe.RunStatusTimestamps{
			PlanQueueableAt: TimeNow(),
		},
		Plan:  newPlan(),
		Apply: newApply(),
	}

	ws, err := f.WorkspaceService.GetByID(opts.Workspace.ID)
	if err != nil {
		return nil, err
	}
	run.Workspace = ws

	cv, err := f.getConfigurationVersion(opts)
	if err != nil {
		return nil, err
	}
	run.ConfigurationVersion = cv

	if opts.IsDestroy != nil {
		run.IsDestroy = *opts.IsDestroy
	}

	if opts.Message != nil {
		run.Message = *opts.Message
	}

	if opts.Refresh != nil {
		run.Refresh = *opts.Refresh
	}

	return &run, nil
}

func (f *RunFactory) getConfigurationVersion(opts *tfe.RunCreateOptions) (*ConfigurationVersion, error) {
	// Unless CV ID provided, get workspace's latest CV
	if opts.ConfigurationVersion != nil {
		return f.ConfigurationVersionService.Get(opts.ConfigurationVersion.ID)
	}
	return f.ConfigurationVersionService.GetLatest(opts.Workspace.ID)
}
