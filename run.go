package ots

import (
	"errors"
	"fmt"
	"strings"
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
	ExternalID string `gorm:"uniqueIndex"`
	InternalID uint   `gorm:"primaryKey;column:id"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ForceCancelAvailableAt time.Time
	IsDestroy              bool
	Message                string
	Permissions            *tfe.RunPermissions `gorm:"embedded;embeddedPrefix:permission_"`
	PositionInQueue        int
	Refresh                bool
	RefreshOnly            bool
	Status                 tfe.RunStatus
	StatusTimestamps       *tfe.RunStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	ReplaceAddrs         []string `gorm:"-"`
	InternalReplaceAddrs string   `gorm:"column:replace_addrs"`

	TargetAddrs         []string `gorm:"-"`
	InternalTargetAddrs string   `gorm:"column:target_addrs"`

	CreatedBy *tfe.User `gorm:"-"`

	// Run has one plan
	Plan *Plan

	// Run has one apply
	Apply *Apply

	WorkspaceID uint
	Workspace   *Workspace

	// Run belongs to a configuration version
	ConfigurationVersionID uint
	ConfigurationVersion   *ConfigurationVersion
}

func (run *Run) Unwrap(tx *gorm.DB) (err error) {
	run.ReplaceAddrs = strings.Split(run.InternalReplaceAddrs, ",")
	run.TargetAddrs = strings.Split(run.InternalTargetAddrs, ",")
	run.CreatedBy = &tfe.User{
		ID:       DefaultUserID,
		Username: DefaultUsername,
	}

	return
}

func (run *Run) Wrap(tx *gorm.DB) (err error) {
	run.InternalReplaceAddrs = strings.Join(run.ReplaceAddrs, ",")
	run.InternalTargetAddrs = strings.Join(run.TargetAddrs, ",")
	return
}

func (run *Run) AfterFind(tx *gorm.DB) (err error) { run.Unwrap(tx); return }

func (run *Run) BeforeSave(tx *gorm.DB) (err error) { run.Wrap(tx); return }
func (run *Run) AfterSave(tx *gorm.DB) (err error)  { run.Unwrap(tx); return }

func (r *Run) DTO() interface{} {
	return &tfe.Run{
		ID:                     r.ExternalID,
		Actions:                r.Actions(),
		CreatedAt:              r.CreatedAt,
		ForceCancelAvailableAt: r.ForceCancelAvailableAt,
		HasChanges:             false,
		IsDestroy:              r.IsDestroy,
		Message:                r.Message,
		Permissions:            r.Permissions,
		PositionInQueue:        0,
		Refresh:                r.Refresh,
		RefreshOnly:            r.RefreshOnly,
		ReplaceAddrs:           r.ReplaceAddrs,
		Source:                 DefaultConfigurationSource,
		Status:                 r.Status,
		StatusTimestamps:       r.StatusTimestamps,
		TargetAddrs:            r.TargetAddrs,

		// Relations
		Apply:                r.Apply.DTO().(*tfe.Apply),
		ConfigurationVersion: r.ConfigurationVersion.DTO().(*tfe.ConfigurationVersion),
		CreatedBy:            r.CreatedBy,
		Plan:                 r.Plan.DTO().(*tfe.Plan),
		Workspace:            r.Workspace.DTO().(*tfe.Workspace),
	}
}

func (rl *RunList) DTO() interface{} {
	l := &tfe.RunList{
		Pagination: rl.Pagination,
	}
	for _, item := range rl.Items {
		l.Items = append(l.Items, item.DTO().(*tfe.Run))
	}

	return l
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

func (r *Run) QueueApply() {
	r.Status = tfe.RunApplyQueued
	r.StatusTimestamps.ApplyQueuedAt = time.Now()

	r.Apply.Status = tfe.ApplyQueued
	r.Apply.StatusTimestamps.QueuedAt = time.Now()
}

// PlanFinished updates the state of a run to reflect its plan having finished
func (r *Run) FinishPlan() error {
	if r.ConfigurationVersion.Speculative {
		r.Status = tfe.RunPlannedAndFinished
		r.StatusTimestamps.PlannedAndFinishedAt = time.Now()
	} else {
		r.Status = tfe.RunPlanned
		r.StatusTimestamps.PlannedAt = time.Now()
	}

	r.Plan.UpdateStatus(tfe.PlanFinished)

	return nil
}

// ApplyFinished updates the state of a run to reflect its plan having finished
func (r *Run) FinishApply() error {
	r.Status = tfe.RunApplied
	r.StatusTimestamps.AppliedAt = time.Now()

	r.Apply.UpdateStatus(tfe.ApplyFinished)

	return nil
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.IsDiscardable() {
		return ErrRunDiscardNotAllowed
	}

	r.Status = tfe.RunDiscarded
	r.StatusTimestamps.DiscardedAt = time.Now()

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

func (f *RunFactory) NewRun(opts *tfe.RunCreateOptions) (*Run, error) {
	if opts.Workspace == nil {
		return nil, errors.New("workspace is required")
	}

	run := Run{
		ExternalID: NewRunID(),
		Permissions: &tfe.RunPermissions{
			CanForceCancel: true,
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
	run.Workspace = ws
	run.WorkspaceID = ws.InternalID

	cv, err := f.getConfigurationVersion(opts)
	if err != nil {
		return nil, err
	}
	run.ConfigurationVersion = cv
	run.ConfigurationVersionID = cv.InternalID

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
