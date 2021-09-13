package ots

import (
	"errors"
	"fmt"
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
	Apply(id string, opts *tfe.RunApplyOptions) error
	Discard(id string, opts *tfe.RunDiscardOptions) error
	Cancel(id string, opts *tfe.RunCancelOptions) error
	ForceCancel(id string, opts *tfe.RunForceCancelOptions) error
	EnqueuePlan(id string) error
	UpdateStatus(id string, status tfe.RunStatus) (*Run, error)
	GetPlanLogs(id string, opts GetLogOptions) ([]byte, error)
	UploadPlanLogs(id string, logs []byte, opts AppendLogOptions) error
	GetApplyLogs(id string, opts GetLogOptions) ([]byte, error)
	UploadApplyLogs(id string, logs []byte, opts AppendLogOptions) error
	FinishPlan(id string, opts PlanFinishOptions) (*Run, error)
	FinishApply(id string, opts ApplyFinishOptions) (*Run, error)
	GetPlanJSON(id string) ([]byte, error)
	GetPlanFile(id string) ([]byte, error)
	UploadPlan(runID string, plan []byte, json bool) error
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

// RunGetOptions are options for retrieving a single Run. Either ID or ApplyID
// or PlanID must be specfiied.
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
func (r *Run) FinishPlan(opts PlanFinishOptions) {
	r.Plan.ResourceAdditions = opts.ResourceAdditions
	r.Plan.ResourceChanges = opts.ResourceChanges
	r.Plan.ResourceDestructions = opts.ResourceDestructions

	r.UpdateStatus(tfe.RunPlanned)
}

// FinishApply updates the state of a run to reflect its plan having finished
func (r *Run) FinishApply(opts ApplyFinishOptions) {
	r.Apply.ResourceAdditions = opts.ResourceAdditions
	r.Apply.ResourceChanges = opts.ResourceChanges
	r.Apply.ResourceDestructions = opts.ResourceDestructions

	r.UpdateStatus(tfe.RunApplied)
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.IsDiscardable() {
		return ErrRunDiscardNotAllowed
	}

	r.UpdateStatus(tfe.RunDiscarded)

	return nil
}

// Cancel updates the state of a run to reflect a cancel request having
// been issued.
func (r *Run) Cancel() error {
	if !r.IsCancelable() {
		return ErrRunCancelNotAllowed
	}

	// Run can be forcefully cancelled after a cool-off period of ten seconds
	r.ForceCancelAvailableAt = time.Now().Add(10 * time.Second)

	r.UpdateStatus(tfe.RunCanceled)

	return nil
}

// ForceCancel updates the state of a run to reflect it having been forcefully
// cancelled.
func (r *Run) ForceCancel() error {
	if !r.IsForceCancelable() {
		return ErrRunForceCancelNotAllowed
	}

	r.StatusTimestamps.ForceCanceledAt = TimeNow()

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

func (r *Run) IsSpeculative() bool {
	return r.ConfigurationVersion.Speculative
}

// UpdateStatus updates the status of the run as well as its plan and apply
func (r *Run) UpdateStatus(status tfe.RunStatus) error {
	switch status {
	case tfe.RunPending:
		r.Plan.UpdateStatus(tfe.PlanPending)
	case tfe.RunPlanQueued:
		r.Plan.UpdateStatus(tfe.PlanQueued)
	case tfe.RunPlanning:
		r.Plan.UpdateStatus(tfe.PlanRunning)
	case tfe.RunPlanned, tfe.RunPlannedAndFinished:
		r.Plan.UpdateStatus(tfe.PlanFinished)
	case tfe.RunApplyQueued:
		r.Apply.UpdateStatus(tfe.ApplyQueued)
	case tfe.RunApplying:
		r.Apply.UpdateStatus(tfe.ApplyRunning)
	case tfe.RunApplied:
		r.Apply.UpdateStatus(tfe.ApplyFinished)
	case tfe.RunErrored:
		switch r.Status {
		case tfe.RunPlanning:
			r.Plan.UpdateStatus(tfe.PlanErrored)
		case tfe.RunApplying:
			r.Apply.UpdateStatus(tfe.ApplyErrored)
		}
	case tfe.RunCanceled:
		switch r.Status {
		case tfe.RunPlanQueued, tfe.RunPlanning:
			r.Plan.UpdateStatus(tfe.PlanCanceled)
		case tfe.RunApplyQueued, tfe.RunApplying:
			r.Apply.UpdateStatus(tfe.ApplyCanceled)
		}
	}

	r.Status = status
	r.setTimestamp(status)

	// TODO: determine when tfe.ApplyUnreachable is applicable and set
	// accordingly

	return nil
}

func (r *Run) setTimestamp(status tfe.RunStatus) {
	switch status {
	case tfe.RunPending:
		r.StatusTimestamps.PlanQueueableAt = TimeNow()
	case tfe.RunPlanQueued:
		r.StatusTimestamps.PlanQueuedAt = TimeNow()
	case tfe.RunPlanning:
		r.StatusTimestamps.PlanningAt = TimeNow()
	case tfe.RunPlanned:
		r.StatusTimestamps.PlannedAt = TimeNow()
	case tfe.RunPlannedAndFinished:
		r.StatusTimestamps.PlannedAndFinishedAt = TimeNow()
	case tfe.RunApplyQueued:
		r.StatusTimestamps.ApplyQueuedAt = TimeNow()
	case tfe.RunApplying:
		r.StatusTimestamps.ApplyingAt = TimeNow()
	case tfe.RunApplied:
		r.StatusTimestamps.AppliedAt = TimeNow()
	case tfe.RunErrored:
		r.StatusTimestamps.ErroredAt = TimeNow()
	case tfe.RunCanceled:
		r.StatusTimestamps.CanceledAt = TimeNow()
	case tfe.RunDiscarded:
		r.StatusTimestamps.DiscardedAt = TimeNow()
	}
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
		Refresh:          DefaultRefresh,
		ReplaceAddrs:     opts.ReplaceAddrs,
		TargetAddrs:      opts.TargetAddrs,
		StatusTimestamps: &tfe.RunStatusTimestamps{},
		Plan:             newPlan(),
		Apply:            newApply(),
	}

	run.UpdateStatus(tfe.RunPending)

	ws, err := f.WorkspaceService.Get(WorkspaceSpecifier{ID: &opts.Workspace.ID})
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
