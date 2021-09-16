package ots

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// Phase implementations represent the phases that make up a run: a plan and an
// apply.
type Phase interface {
	GetLogsBlobID() string
	Do(*Run, *Environment) error
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
	GetPlanLogs(id string, opts GetChunkOptions) ([]byte, error)
	UploadLogs(id string, logs []byte, opts PutChunkOptions) error
	GetApplyLogs(id string, opts GetChunkOptions) ([]byte, error)
	GetPlanJSON(id string) ([]byte, error)
	GetPlanFile(id string) ([]byte, error)
	UploadPlan(runID string, plan []byte, json bool) error

	JobService
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

func (r *Run) GetID() string {
	return r.ID
}

func (r *Run) GetStatus() string {
	return string(r.Status)
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.IsDiscardable() {
		return ErrRunDiscardNotAllowed
	}

	r.UpdateStatus(tfe.RunDiscarded)

	return nil
}

// Cancel run.
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

// ActivePhase retrieves the currently active phase
func (r *Run) ActivePhase() (Phase, error) {
	switch r.Status {
	case tfe.RunPlanning:
		return r.Plan, nil
	case tfe.RunApplying:
		return r.Apply, nil
	default:
		return nil, fmt.Errorf("invalid run status: %s", r.Status)
	}
}

// Start starts a run phase.
func (r *Run) Start() error {
	switch r.Status {
	case tfe.RunPlanQueued:
		r.UpdateStatus(tfe.RunPlanning)
	case tfe.RunApplyQueued:
		r.UpdateStatus(tfe.RunApplying)
	default:
		return fmt.Errorf("run cannot be started: invalid status: %s", r.Status)
	}

	return nil
}

// Finish updates the run to reflect the current phase having finished. An event
// is emitted reflecting the run's new status.
func (r *Run) Finish(bs BlobStore) (*Event, error) {
	if r.Status == tfe.RunApplying {
		r.UpdateStatus(tfe.RunApplied)

		if err := r.Apply.UpdateResources(bs); err != nil {
			return nil, err
		}

		return &Event{Payload: r, Type: RunApplied}, nil
	}

	// Only remaining valid status is planning
	if r.Status != tfe.RunPlanning {
		return nil, fmt.Errorf("run cannot be finished: invalid status: %s", r.Status)
	}

	if err := r.Plan.UpdateResources(bs); err != nil {
		return nil, err
	}

	// Speculative plan, proceed no further
	if r.ConfigurationVersion.Speculative {
		r.UpdateStatus(tfe.RunPlannedAndFinished)
		return &Event{Payload: r, Type: RunPlannedAndFinished}, nil
	}

	r.UpdateStatus(tfe.RunPlanned)

	if r.Workspace.AutoApply {
		r.UpdateStatus(tfe.RunApplyQueued)
		return &Event{Type: ApplyQueued, Payload: r}, nil
	}

	return &Event{Payload: r, Type: RunPlanned}, nil
}

// UpdateStatus updates the status of the run as well as its plan and apply
func (r *Run) UpdateStatus(status tfe.RunStatus) {
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

func (r *Run) Do(env *Environment) error {
	if err := env.RunFunc(r.downloadConfig); err != nil {
		return err
	}

	if err := env.RunFunc(deleteBackendConfigFromDirectory); err != nil {
		return err
	}

	if err := env.RunFunc(r.downloadState); err != nil {
		return err
	}

	if err := env.RunCLI("terraform", "init", "-no-color"); err != nil {
		return err
	}

	phase, err := r.ActivePhase()
	if err != nil {
		return err
	}

	if err := phase.Do(r, env); err != nil {
		return err
	}

	return nil
}

func (r *Run) downloadConfig(ctx context.Context, env *Environment) error {
	// Download config
	cv, err := env.ConfigurationVersionService.Download(r.ConfigurationVersion.ID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}

	// Decompress and untar config
	if err := Unpack(bytes.NewBuffer(cv), env.Path); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}

	return nil
}

// downloadState downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func (r *Run) downloadState(ctx context.Context, env *Environment) error {
	state, err := env.StateVersionService.Current(r.Workspace.ID)
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	statefile, err := env.StateVersionService.Download(state.ID)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(env.Path, LocalStateFilename), statefile, 0644); err != nil {
		return err
	}

	return nil
}

func (r *Run) uploadPlan(ctx context.Context, env *Environment) error {
	file, err := os.ReadFile(filepath.Join(env.Path, PlanFilename))
	if err != nil {
		return err
	}

	if err := env.RunService.UploadPlan(r.ID, file, false); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Run) uploadJSONPlan(ctx context.Context, env *Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.Path, JSONPlanFilename))
	if err != nil {
		return err
	}

	if err := env.RunService.UploadPlan(r.ID, jsonFile, true); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}

	return nil
}

func (r *Run) downloadPlanFile(ctx context.Context, env *Environment) error {
	plan, err := env.RunService.GetPlanFile(r.ID)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(env.Path, PlanFilename), plan, 0644)
}

// uploadState reads, parses, and uploads state
func (r *Run) uploadState(ctx context.Context, env *Environment) error {
	stateFile, err := os.ReadFile(filepath.Join(env.Path, LocalStateFilename))
	if err != nil {
		return err
	}

	state, err := Parse(stateFile)
	if err != nil {
		return err
	}

	_, err = env.StateVersionService.Create(r.Workspace.ID, tfe.StateVersionCreateOptions{
		State:   String(base64.StdEncoding.EncodeToString(stateFile)),
		MD5:     String(fmt.Sprintf("%x", md5.Sum(stateFile))),
		Lineage: &state.Lineage,
		Serial:  Int64(state.Serial),
	})
	if err != nil {
		return err
	}

	return nil
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
