package otf

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
)

const (
	// DefaultRefresh specifies that the state be refreshed prior to running a
	// plan
	DefaultRefresh = true

	// List all available run statuses supported in OTF.
	RunApplied            RunStatus = "applied"
	RunApplyQueued        RunStatus = "apply_queued"
	RunApplying           RunStatus = "applying"
	RunCanceled           RunStatus = "canceled"
	RunForceCanceled      RunStatus = "force_canceled"
	RunConfirmed          RunStatus = "confirmed"
	RunDiscarded          RunStatus = "discarded"
	RunErrored            RunStatus = "errored"
	RunPending            RunStatus = "pending"
	RunPlanQueued         RunStatus = "plan_queued"
	RunPlanned            RunStatus = "planned"
	RunPlannedAndFinished RunStatus = "planned_and_finished"
	RunPlanning           RunStatus = "planning"
)

var (
	ErrRunDiscardNotAllowed     = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed      = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")

	ErrInvalidRunGetOptions = errors.New("invalid run get options")

	// ActiveRunStatuses are those run statuses that deem a run to be active.
	// There can only be at most one active run for a workspace.
	ActiveRunStatuses = []RunStatus{
		RunApplyQueued,
		RunApplying,
		RunConfirmed,
		RunPlanQueued,
		RunPlanned,
		RunPlanning,
	}

	IncompleteRunStatuses = append(ActiveRunStatuses, RunPending)

	CompletedRunStatuses = []RunStatus{
		RunApplied,
		RunErrored,
		RunDiscarded,
		RunCanceled,
		RunForceCanceled,
	}
)

// RunStatus represents a run state.
type RunStatus string

func (r RunStatus) String() string { return string(r) }

type Run struct {
	ID string `jsonapi:"primary,runs" json:"run_id"`

	Timestamps

	IsDestroy        bool
	Message          string
	PositionInQueue  int
	Refresh          bool
	RefreshOnly      bool
	status           RunStatus
	statusTimestamps []RunStatusTimestamp `json:"run_status_timestamps"`
	ReplaceAddrs     []string
	TargetAddrs      []string

	// Relations
	Plan                 *Plan
	Apply                *Apply
	Workspace            *Workspace            `json:"workspace"`
	ConfigurationVersion *ConfigurationVersion `json:"configuration_version"`

	// Job is the current job the run is performing
	Job
}

type RunStatusTimestamp struct {
	Status    RunStatus `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// RunService implementations allow interactions with runs
type RunService interface {
	// Create a new run with the given options.
	Create(ctx context.Context, opts RunCreateOptions) (*Run, error)

	// Get retrieves a run with the given ID.
	Get(ctx context.Context, id string) (*Run, error)

	// List lists runs according to the given options.
	List(ctx context.Context, opts RunListOptions) (*RunList, error)

	// Delete deletes a run with the given ID.
	Delete(ctx context.Context, id string) error

	// Apply a run with the given ID.
	Apply(ctx context.Context, id string, opts RunApplyOptions) error

	// Discard discards a run with the given ID.
	Discard(ctx context.Context, id string, opts RunDiscardOptions) error
	Cancel(ctx context.Context, id string, opts RunCancelOptions) error
	ForceCancel(ctx context.Context, id string, opts RunForceCancelOptions) error
	// Start a run.
	Start(ctx context.Context, id string) (*Run, error)
	// GetPlanFile retrieves a run's plan file with the requested format.
	GetPlanFile(ctx context.Context, spec RunGetOptions, format PlanFormat) ([]byte, error)
	// UploadPlanFile saves a run's plan file with the requested format.
	UploadPlanFile(ctx context.Context, runID string, plan []byte, format PlanFormat) error
}

// RunCreateOptions represents the options for creating a new run.
type RunCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,runs"`

	// Specifies if this plan is a destroy plan, which will destroy all
	// provisioned resources.
	IsDestroy *bool `jsonapi:"attr,is-destroy,omitempty"`

	// Refresh determines if the run should
	// update the state prior to checking for differences
	Refresh *bool `jsonapi:"attr,refresh,omitempty"`

	// RefreshOnly determines whether the run should ignore config changes
	// and refresh the state only
	RefreshOnly *bool `jsonapi:"attr,refresh-only,omitempty"`

	// Specifies the message to be associated with this run.
	Message *string `jsonapi:"attr,message,omitempty"`

	// Specifies the configuration version to use for this run. If the
	// configuration version object is omitted, the run will be created using the
	// workspace's latest configuration version.
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relation,configuration-version"`

	// Specifies the workspace where the run will be executed.
	Workspace *Workspace `jsonapi:"relation,workspace"`

	// If non-empty, requests that Terraform should create a plan including
	// actions only for the given objects (specified using resource address
	// syntax) and the objects they depend on.
	//
	// This capability is provided for exceptional circumstances only, such as
	// recovering from mistakes or working around existing Terraform
	// limitations. Terraform will generally mention the -target command line
	// option in its error messages describing situations where setting this
	// argument may be appropriate. This argument should not be used as part
	// of routine workflow and Terraform will emit warnings reminding about
	// this whenever this property is set.
	TargetAddrs []string `jsonapi:"attr,target-addrs,omitempty"`

	// If non-empty, requests that Terraform create a plan that replaces
	// (destroys and then re-creates) the objects specified by the given
	// resource addresses.
	ReplaceAddrs []string `jsonapi:"attr,replace-addrs,omitempty"`
}

// RunApplyOptions represents the options for applying a run.
type RunApplyOptions struct {
	// An optional comment about the run.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// RunCancelOptions represents the options for canceling a run.
type RunCancelOptions struct {
	// An optional explanation for why the run was canceled.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// RunForceCancelOptions represents the options for force-canceling a run.
type RunForceCancelOptions struct {
	// An optional comment explaining the reason for the force-cancel.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// RunDiscardOptions represents the options for discarding a run.
type RunDiscardOptions struct {
	// An optional explanation for why the run was discarded.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// RunStore implementations persist Run objects.
type RunStore interface {
	Create(run *Run) error
	Get(opts RunGetOptions) (*Run, error)
	SetPlanFile(id string, file []byte, format PlanFormat) error
	GetPlanFile(id string, format PlanFormat) ([]byte, error)
	List(opts RunListOptions) (*RunList, error)
	// UpdateStatus updates the run's status, providing a func with which to
	// perform updates in a transaction.
	UpdateStatus(opts RunGetOptions, fn func(*Run) error) (*Run, error)
	CreatePlanReport(planID string, report ResourceReport) error
	CreateApplyReport(applyID string, report ResourceReport) error
	Delete(id string) error
}

type RunStatusUpdates struct {
	RunStatus   RunStatus
	PlanStatus  *PlanStatus
	ApplyStatus *ApplyStatus
}

// RunStatusUpdater persists updates to run statuses and returns timestamps of
// when they were persisted.
type RunStatusUpdater interface {
	UpdateRunStatus(ctx context.Context, status RunStatus) (*Run, error)
	UpdatePlanStatus(ctx context.Context, status PlanStatus) (*Plan, error)
	UpdateApplyStatus(ctx context.Context, status ApplyStatus) (*Apply, error)
}

// RunList represents a list of runs.
type RunList struct {
	*Pagination
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

func (o *RunGetOptions) String() string {
	if o.ID != nil {
		return *o.ID
	} else if o.PlanID != nil {
		return *o.PlanID
	} else if o.ApplyID != nil {
		return *o.ApplyID
	} else {
		panic("no ID specified")
	}
}

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	ListOptions

	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include"`

	// Filter by run statuses (with an implicit OR condition)
	Statuses []RunStatus

	// Filter by workspace ID
	WorkspaceID *string `schema:"workspace_id"`

	// Filter by organization and workspace name. Mutually exclusive with
	// WorkspaceID.
	OrganizationName *string `schema:"organization_name"`
	WorkspaceName    *string `schema:"workspace_name"`
}

func (r *Run) String() string { return r.ID }

func (o RunCreateOptions) Valid() error {
	if o.Workspace == nil {
		return errors.New("workspace is required")
	}
	return nil
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.IsDiscardable() {
		return ErrRunDiscardNotAllowed
	}

	r.UpdateStatus(RunDiscarded)

	return nil
}

// Cancel run.
func (r *Run) Cancel() error {
	if !r.IsCancelable() {
		return ErrRunCancelNotAllowed
	}

	r.UpdateStatus(RunCanceled)

	return nil
}

func (r *Run) ForceCancelAvailableAt() time.Time {
	if r.status != RunCanceled {
		return time.Time{}
	}

	canceledAt, found := r.FindRunStatusTimestamp(r.status)
	if !found {
		panic("no corresponding timestamp found for canceled status")
	}

	// Run can be forcefully cancelled after a cool-off period of ten seconds
	return canceledAt.Add(10 * time.Second)
}

// ForceCancel updates the state of a run to reflect it having been forcefully
// cancelled.
func (r *Run) ForceCancel() error {
	if !r.IsForceCancelable() {
		return ErrRunForceCancelNotAllowed
	}

	return r.UpdateStatus(RunForceCanceled)
}

// IsCancelable determines whether run can be cancelled.
func (r *Run) IsCancelable() bool {
	switch r.Status() {
	case RunPending, RunPlanQueued, RunPlanning, RunApplyQueued, RunApplying:
		return true
	default:
		return false
	}
}

// IsConfirmable determines whether run can be confirmed.
func (r *Run) IsConfirmable() bool {
	switch r.Status() {
	case RunPlanned:
		return true
	default:
		return false
	}
}

// IsDiscardable determines whether run can be discarded.
func (r *Run) IsDiscardable() bool {
	switch r.Status() {
	case RunPending, RunPlanned:
		return true
	default:
		return false
	}
}

// IsForceCancelable determines whether a run can be forcibly cancelled.
func (r *Run) IsForceCancelable() bool {
	availAt := r.ForceCancelAvailableAt()

	if availAt.IsZero() {
		return false
	}

	return time.Now().After(availAt)
}

// IsActive determines whether run is currently the active run on a workspace,
// i.e. it is neither finished nor pending
func (r *Run) IsActive() bool {
	if r.IsDone() || r.Status() == RunPending {
		return false
	}
	return true
}

// IsDone determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (r *Run) IsDone() bool {
	switch r.Status() {
	case RunApplied, RunPlannedAndFinished, RunDiscarded, RunCanceled, RunErrored:
		return true
	default:
		return false
	}
}

func (r *Run) IsSpeculative() bool {
	return r.ConfigurationVersion.Speculative
}

func (r *Run) ApplyRun() error {
	return r.UpdateStatus(RunApplyQueued)
}

func (r *Run) EnqueuePlan() error {
	return r.UpdateStatus(RunPlanQueued)
}

// UpdateStatus updates the status of the run as well as its plan and apply
func (r *Run) UpdateStatus(status RunStatus) error {
	switch status {
	case RunPending:
		r.Plan.status = PlanPending
	case RunPlanQueued:
		r.Plan.status = PlanQueued
	case RunPlanning:
		r.Plan.status = PlanRunning
	case RunPlanned, RunPlannedAndFinished:
		r.Plan.status = PlanFinished
	case RunApplyQueued:
		r.Apply.status = ApplyQueued
	case RunApplying:
		r.Apply.status = ApplyRunning
	case RunApplied:
		r.Apply.status = ApplyFinished
	case RunErrored:
		switch r.Status() {
		case RunPlanning:
			r.Plan.status = PlanErrored
		case RunApplying:
			r.Apply.status = ApplyErrored
		}
	case RunCanceled:
		switch r.Status() {
		case RunPlanQueued, RunPlanning:
			r.Plan.status = PlanCanceled
		case RunApplyQueued, RunApplying:
			r.Apply.status = ApplyCanceled
		}
	}

	r.status = status

	// set job reflecting new status
	r.setJob()

	// TODO: determine when ApplyUnreachable is applicable and set accordingly

	return nil
}

// setupEnv invokes the necessary steps before a plan or apply can proceed.
func (r *Run) setupEnv(env Environment) error {
	if err := env.RunFunc(r.downloadConfig); err != nil {
		return err
	}

	err := env.RunFunc(func(ctx context.Context, env Environment) error {
		return deleteBackendConfigFromDirectory(ctx, env.GetPath())
	})
	if err != nil {
		return err
	}

	if err := env.RunFunc(r.downloadState); err != nil {
		return err
	}

	if err := env.RunCLI("terraform", "init"); err != nil {
		return fmt.Errorf("running terraform init: %w", err)
	}

	return nil
}

func (r *Run) downloadConfig(ctx context.Context, env Environment) error {
	// Download config
	cv, err := env.GetConfigurationVersionService().Download(r.ConfigurationVersion.ID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}

	// Decompress and untar config
	if err := Unpack(bytes.NewBuffer(cv), env.GetPath()); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}

	return nil
}

// downloadState downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func (r *Run) downloadState(ctx context.Context, env Environment) error {
	state, err := env.GetStateVersionService().Current(r.Workspace.ID)
	if errors.Is(err, ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("retrieving current state version: %w", err)
	}

	statefile, err := env.GetStateVersionService().Download(state.ID)
	if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}

	if err := os.WriteFile(filepath.Join(env.GetPath(), LocalStateFilename), statefile, 0644); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}

	return nil
}

func (r *Run) uploadPlan(ctx context.Context, env Environment) error {
	file, err := os.ReadFile(filepath.Join(env.GetPath(), PlanFilename))
	if err != nil {
		return err
	}

	if err := env.GetRunService().UploadPlanFile(ctx, r.ID, file, PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Run) uploadJSONPlan(ctx context.Context, env Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.GetPath(), JSONPlanFilename))
	if err != nil {
		return err
	}

	if err := env.GetRunService().UploadPlanFile(ctx, r.ID, jsonFile, PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}

	return nil
}

func (r *Run) downloadPlanFile(ctx context.Context, env Environment) error {
	plan, err := env.GetRunService().GetPlanFile(ctx, RunGetOptions{ID: &r.ID}, PlanFormatBinary)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(env.GetPath(), PlanFilename), plan, 0644)
}

// uploadState reads, parses, and uploads terraform state
func (r *Run) uploadState(ctx context.Context, env Environment) error {
	stateFile, err := os.ReadFile(filepath.Join(env.GetPath(), LocalStateFilename))
	if err != nil {
		return err
	}

	state, err := Parse(stateFile)
	if err != nil {
		return err
	}

	_, err = env.GetStateVersionService().Create(r.Workspace.ID, StateVersionCreateOptions{
		State:   String(base64.StdEncoding.EncodeToString(stateFile)),
		MD5:     String(fmt.Sprintf("%x", md5.Sum(stateFile))),
		Lineage: &state.Lineage,
		Serial:  Int64(state.Serial),
		Run:     r,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Run) Status() RunStatus                      { return r.status }
func (r *Run) StatusTimestamps() []RunStatusTimestamp { return r.statusTimestamps }

func (r *Run) AddStatusTimestamp(status RunStatus, timestamp time.Time) {
	r.statusTimestamps = append(r.statusTimestamps, RunStatusTimestamp{
		Status:    status,
		Timestamp: timestamp,
	})
}

func (r *Run) FindRunStatusTimestamp(status RunStatus) (time.Time, bool) {
	for _, rst := range r.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, true
		}
	}
	return time.Time{}, false
}

// Set appropriate job for run
func (r *Run) setJob() {
	switch r.Status() {
	case RunPlanQueued, RunPlanning:
		r.Job = r.Plan
	case RunApplyQueued, RunApplying:
		r.Job = r.Apply
	default:
		r.Job = &noOp{}
	}
}
func ContainsRunStatus(statuses []RunStatus, status RunStatus) bool {
	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}
