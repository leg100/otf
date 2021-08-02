package ots

import (
	"errors"
	"fmt"
	"time"

	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

const (
	DefaultRefresh = true
)

var (
	ErrRunDiscardNotAllowed     = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed      = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")

	ErrInvalidRunGetOptions = errors.New("invalid run get options")
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

	Plan *Plan

	Apply *Apply

	WorkspaceID uint
	Workspace   *Workspace

	ConfigurationVersionID uint
	ConfigurationVersion   *ConfigurationVersion
}

type RunFactory struct {
	ConfigurationVersionService ConfigurationVersionService
	WorkspaceService            WorkspaceService
}

type RunService interface {
	Create(opts *tfe.RunCreateOptions) (*Run, error)
	Get(id string) (*Run, error)
	List(workspaceID string, opts tfe.RunListOptions) (*RunList, error)
	GetQueued(opts tfe.RunListOptions) (*RunList, error)
	Apply(id string, opts *tfe.RunApplyOptions) error
	Discard(id string, opts *tfe.RunDiscardOptions) error
	Cancel(id string, opts *tfe.RunCancelOptions) error
	ForceCancel(id string, opts *tfe.RunForceCancelOptions) error
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

type RunStore interface {
	Create(run *Run) (*Run, error)
	Get(opts RunGetOptions) (*Run, error)
	List(opts RunListOptions) (*RunList, error)
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

// RunListOptions are options for paginating and filtering the list of runs to
// retrieve from the RunRepository ListRuns endpoint
type RunListOptions struct {
	tfe.ListOptions

	// Filter by run statuses (with an implicit OR condition)
	Statuses []tfe.RunStatus

	// Filter by workspace ID
	WorkspaceID *string
}

func NewRunID() string {
	return fmt.Sprintf("run-%s", GenerateRandomString(16))
}

// PlanFinished updates the state of a run to reflect its plan having finished
func (r *Run) FinishPlan(opts PlanFinishOptions) error {
	r.Plan.ResourceAdditions = opts.ResourceAdditions
	r.Plan.ResourceChanges = opts.ResourceChanges
	r.Plan.ResourceDestructions = opts.ResourceDestructions
	r.Plan.Plan = opts.Plan
	r.Plan.PlanJSON = opts.PlanJSON

	r.UpdatePlanStatus(tfe.PlanFinished)

	return nil
}

// ApplyFinished updates the state of a run to reflect its plan having finished
func (r *Run) FinishApply(opts ApplyFinishOptions) error {
	r.Apply.ResourceAdditions = opts.ResourceAdditions
	r.Apply.ResourceChanges = opts.ResourceChanges
	r.Apply.ResourceDestructions = opts.ResourceDestructions

	r.UpdateApplyStatus(tfe.ApplyFinished)

	return nil
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.IsDiscardable() {
		return ErrRunDiscardNotAllowed
	}

	r.UpdateStatus(tfe.RunDiscarded)

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
		r.Plan.StatusTimestamps.CanceledAt = time.Now()
	case tfe.RunApplyQueued, tfe.RunApplying:
		r.Apply.Status = tfe.ApplyCanceled
		r.Apply.StatusTimestamps.CanceledAt = time.Now()
	}

	// Put run into a cancel state
	r.Status = tfe.RunCanceled
	r.StatusTimestamps.CanceledAt = time.Now()

	return nil
}

func (r *Run) Actions() *tfe.RunActions {
	return &tfe.RunActions{
		IsCancelable:      r.IsCancelable(),
		IsConfirmable:     r.IsConfirmable(),
		IsForceCancelable: r.IsForceCancelable(),
		IsDiscardable:     r.IsDiscardable(),
	}
}

func (r *Run) IsCancelable() bool {
	switch r.Status {
	case tfe.RunPending, tfe.RunPlanQueued, tfe.RunPlanning, tfe.RunApplyQueued, tfe.RunApplying:
		return true
	default:
		return false
	}
}

func (r *Run) IsConfirmable() bool {
	switch r.Status {
	case tfe.RunPlanned:
		return true
	default:
		return false
	}
}

func (r *Run) IsDiscardable() bool {
	switch r.Status {
	case tfe.RunPending, tfe.RunPolicyChecked, tfe.RunPolicyOverride, tfe.RunPlanned:
		return true
	default:
		return false
	}
}

func (r *Run) IsForceCancelable() bool {
	return r.IsCancelable() && !r.ForceCancelAvailableAt.IsZero() && time.Now().After(r.ForceCancelAvailableAt)
}

// UpdateStatus updates the status of the run.
func (r *Run) UpdateStatus(status tfe.RunStatus) {
	timestamps := &tfe.RunStatusTimestamps{}
	if r.StatusTimestamps != nil {
		timestamps = r.StatusTimestamps
	}

	switch status {
	case tfe.RunDiscarded:
		timestamps.DiscardedAt = time.Now()
		r.UpdateApplyStatus(tfe.ApplyUnreachable)
	case tfe.RunPlanQueued:
		timestamps.PlanQueuedAt = time.Now()
	case tfe.RunApplyQueued:
		timestamps.ApplyQueuedAt = time.Now()
	case tfe.RunApplied:
		timestamps.AppliedAt = time.Now()
	case tfe.RunErrored:
		timestamps.ErroredAt = time.Now()
	default:
		// Don't set a status or timestamp
		return
	}

	r.Status = status
	r.StatusTimestamps = timestamps
}

// UpdateStatus updates the status of the plan. It'll also update the
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
		timestamps.QueuedAt = time.Now()
		r.UpdateStatus(tfe.RunPlanQueued)
	case tfe.PlanCanceled:
		timestamps.CanceledAt = time.Now()
	case tfe.PlanErrored:
		timestamps.ErroredAt = time.Now()
		r.UpdateStatus(tfe.RunErrored)
	case tfe.PlanFinished:
		timestamps.FinishedAt = time.Now()

		if r.ConfigurationVersion.Speculative {
			r.Status = tfe.RunPlannedAndFinished
			r.StatusTimestamps.PlannedAndFinishedAt = time.Now()
		} else {
			r.Status = tfe.RunPlanned
			r.StatusTimestamps.PlannedAt = time.Now()
		}
	default:
		// Don't set a timestamp
		return
	}

	r.Plan.Status = status

	// Set timestamps on plan
	r.Plan.StatusTimestamps = timestamps
}

// UpdateStatus updates the status of the apply. It'll also update the
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
		timestamps.FinishedAt = time.Now()
		r.UpdateStatus(tfe.RunApplied)
	case tfe.ApplyRunning:
		timestamps.StartedAt = time.Now()
		r.UpdateStatus(tfe.RunApplying)
	case tfe.ApplyQueued:
		timestamps.QueuedAt = time.Now()
		r.UpdateStatus(tfe.RunApplyQueued)
	case tfe.ApplyCanceled:
		timestamps.CanceledAt = time.Now()
	case tfe.ApplyErrored:
		timestamps.ErroredAt = time.Now()
		r.UpdateStatus(tfe.RunErrored)
	default:
		// Don't set a timestamp
		return
	}

	r.Apply.Status = status

	// Set timestamps on apply
	r.Apply.StatusTimestamps = timestamps
}

func (f *RunFactory) NewRun(opts *tfe.RunCreateOptions) (*Run, error) {
	if opts.Workspace == nil {
		return nil, errors.New("workspace is required")
	}

	run := Run{
		ID: NewRunID(),
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
		Status:       tfe.RunPlanQueued,
		StatusTimestamps: &tfe.RunStatusTimestamps{
			PlanQueueableAt: time.Now(),
		},
		Plan:  newPlan(),
		Apply: newApply(),
	}

	ws, err := f.WorkspaceService.GetByID(opts.Workspace.ID)
	if err != nil {
		return nil, err
	}
	run.Workspace, run.WorkspaceID = ws, ws.Model.ID

	cv, err := f.getConfigurationVersion(opts)
	if err != nil {
		return nil, err
	}
	run.ConfigurationVersion, run.ConfigurationVersionID = cv, cv.Model.ID

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
