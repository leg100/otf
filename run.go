package otf

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	jsonapi "github.com/leg100/otf/http/dto"
)

const (
	// DefaultRefresh specifies that the state be refreshed prior to running a
	// plan
	DefaultRefresh = true
)

var (
	ErrStatusTimestampNotFound  = errors.New("corresponding status timestamp not found")
	ErrRunDiscardNotAllowed     = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed      = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")
	ErrInvalidRunGetOptions     = errors.New("invalid run get options")
	ActiveRun                   = []RunStatus{
		RunApplyQueued,
		RunApplying,
		RunConfirmed,
		RunPlanQueued,
		RunPlanned,
		RunPlanning,
	}
	IncompleteRun = append(ActiveRun, RunPending)
	CompletedRun  = []RunStatus{
		RunApplied,
		RunErrored,
		RunDiscarded,
		RunCanceled,
		RunForceCanceled,
	}
)

type Run struct {
	id                     string
	createdAt              time.Time
	isDestroy              bool
	message                string
	positionInQueue        int
	refresh                bool
	refreshOnly            bool
	autoApply              bool
	speculative            bool
	statusTimestamps       []RunStatusTimestamp
	replaceAddrs           []string
	targetAddrs            []string
	organizationName       string
	workspaceName          string
	workspaceID            string
	configurationVersionID string
	// Relations
	Plan      *Plan
	Apply     *Apply
	workspace *Workspace

	// available run states
	pendingState            *pendingState
	planQueuedState         *planQueuedState
	planningState           *planningState
	plannedState            *plannedState
	canceledState           *canceledState
	discardedState          *discardedState
	erroredState            *erroredState
	plannedAndFinishedState *plannedAndFinishedState
	applyQueuedState        *applyQueuedState
	applyingState           *applyingState
	appliedState            *appliedState
	// current state
	state runState
	// current phase
	phase Phase
}

func (r *Run) ID() string                             { return r.id }
func (r *Run) RunID() string                          { return r.id }
func (r *Run) CreatedAt() time.Time                   { return r.createdAt }
func (r *Run) String() string                         { return r.id }
func (r *Run) IsDestroy() bool                        { return r.isDestroy }
func (r *Run) Message() string                        { return r.message }
func (r *Run) OrganizationName() string               { return r.organizationName }
func (r *Run) Refresh() bool                          { return r.refresh }
func (r *Run) RefreshOnly() bool                      { return r.refreshOnly }
func (r *Run) ReplaceAddrs() []string                 { return r.replaceAddrs }
func (r *Run) TargetAddrs() []string                  { return r.targetAddrs }
func (r *Run) StatusTimestamps() []RunStatusTimestamp { return r.statusTimestamps }
func (r *Run) WorkspaceName() string                  { return r.workspaceName }
func (r *Run) WorkspaceID() string                    { return r.workspaceID }
func (r *Run) Workspace() *Workspace                  { return r.workspace }
func (r *Run) ConfigurationVersionID() string         { return r.configurationVersionID }
func (r *Run) HasChanges() bool                       { return r.Plan.HasChanges() }

func (r *Run) ForceCancelAvailableAt() time.Time {
	if r.Status() != RunCanceled {
		return time.Time{}
	}
	canceledAt, err := r.StatusTimestamp(r.Status())
	if err != nil {
		panic("no corresponding timestamp found for canceled status")
	}
	// Run can be forcefully cancelled after a cool-off period of ten seconds
	return canceledAt.Add(10 * time.Second)
}

// ForceCancel updates the state of a run to reflect it having been forcefully
// cancelled.
func (r *Run) ForceCancel() error {
	if !r.ForceCancelable() {
		return ErrRunForceCancelNotAllowed
	}
	return nil
}

// ForceCancelable determines whether a run can be forcibly cancelled.
func (r *Run) ForceCancelable() bool {
	availAt := r.ForceCancelAvailableAt()
	if availAt.IsZero() {
		return false
	}
	return CurrentTimestamp().After(availAt)
}

func (r *Run) Speculative() bool {
	return r.speculative
}

func (r *Run) StatusTimestamp(status RunStatus) (time.Time, error) {
	for _, rst := range r.statusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, ErrStatusTimestampNotFound
}

// CanLock determines whether requestor can replace run lock
func (r *Run) CanLock(requestor Identity) error {
	if _, ok := requestor.(*Run); ok {
		// run can replace lock held by different run
		return nil
	}
	return ErrWorkspaceAlreadyLocked
}

// CanUnlock determines whether requestor can unlock run lock
func (r *Run) CanUnlock(requestor Identity, force bool) error {
	if force {
		// TODO: only grant admin user force unlock always granted
		return nil
	}
	if _, ok := requestor.(*Run); ok {
		// runs can unlock other run locks
		return nil
	}
	return ErrWorkspaceLockedByDifferentUser
}

// ToJSONAPI assembles a JSON-API DTO.
func (r *Run) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.Run{
		ID: r.ID(),
		Actions: &jsonapi.RunActions{
			IsCancelable:      r.Cancelable(),
			IsConfirmable:     r.Confirmable(),
			IsForceCancelable: r.ForceCancelable(),
			IsDiscardable:     r.Discardable(),
		},
		CreatedAt:              r.CreatedAt(),
		ForceCancelAvailableAt: r.ForceCancelAvailableAt(),
		HasChanges:             r.Plan.HasChanges(),
		IsDestroy:              r.IsDestroy(),
		Message:                r.Message(),
		Permissions: &jsonapi.RunPermissions{
			CanForceCancel:  true,
			CanApply:        true,
			CanCancel:       true,
			CanDiscard:      true,
			CanForceExecute: true,
		},
		PositionInQueue:  0,
		Refresh:          r.Refresh(),
		RefreshOnly:      r.RefreshOnly(),
		ReplaceAddrs:     r.ReplaceAddrs(),
		Source:           DefaultConfigurationSource,
		Status:           string(r.Status()),
		StatusTimestamps: &jsonapi.RunStatusTimestamps{},
		TargetAddrs:      r.TargetAddrs(),
		// Relations
		Apply: r.Apply.ToJSONAPI(req).(*jsonapi.Apply),
		Plan:  r.Plan.ToJSONAPI(req).(*jsonapi.Plan),
		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &jsonapi.User{
			ID:       DefaultUserID,
			Username: DefaultUsername,
		},
		ConfigurationVersion: &jsonapi.ConfigurationVersion{
			ID: r.configurationVersionID,
		},
	}
	if r.workspace != nil {
		dto.Workspace = r.workspace.ToJSONAPI(req).(*jsonapi.Workspace)
	} else {
		dto.Workspace = &jsonapi.Workspace{
			ID: r.workspaceID,
		}
	}

	for _, rst := range r.StatusTimestamps() {
		switch rst.Status {
		case RunPending:
			dto.StatusTimestamps.PlanQueueableAt = &rst.Timestamp
		case RunPlanQueued:
			dto.StatusTimestamps.PlanQueuedAt = &rst.Timestamp
		case RunPlanning:
			dto.StatusTimestamps.PlanningAt = &rst.Timestamp
		case RunPlanned:
			dto.StatusTimestamps.PlannedAt = &rst.Timestamp
		case RunPlannedAndFinished:
			dto.StatusTimestamps.PlannedAndFinishedAt = &rst.Timestamp
		case RunApplyQueued:
			dto.StatusTimestamps.ApplyQueuedAt = &rst.Timestamp
		case RunApplying:
			dto.StatusTimestamps.ApplyingAt = &rst.Timestamp
		case RunApplied:
			dto.StatusTimestamps.AppliedAt = &rst.Timestamp
		case RunErrored:
			dto.StatusTimestamps.ErroredAt = &rst.Timestamp
		case RunCanceled:
			dto.StatusTimestamps.CanceledAt = &rst.Timestamp
		case RunForceCanceled:
			dto.StatusTimestamps.ForceCanceledAt = &rst.Timestamp
		case RunDiscarded:
			dto.StatusTimestamps.DiscardedAt = &rst.Timestamp
		}
	}
	return dto
}

// Start the current phase
func (r *Run) Start() error {
	// if current phase state is not JobQueued then error

	// transition run state to <Phase>ing / JobRunning

	return nil
}

func (r *Run) Enqueue() error {
	if r.state.phaseState != JobPending {
		return ErrRunInvalidStateTransition
	}
	r.transition(RunPlanQueuedState)

	return nil
}

func (r *Run) transition(state runState) {
	r.state = state
	// add status timestamp for both run state and phase state
}

// Finish the current phase
func (r *Run) Finish() error {
	// if current phase state is not JobRunning then error

	// transition run state to <Planned/Applied> / JobFinished

	return nil
}

// setState transitions the run to a new state
func (r *Run) setState(s runState) {
	r.runState = s
	r.statusTimestamps = append(r.statusTimestamps, RunStatusTimestamp{
		Status:    s.Status(),
		Timestamp: CurrentTimestamp(),
	})
}

// setStateFromStatus sets the current run state from a run status string
func (r *Run) setStateFromStatus(status RunStatus) error {
	switch status {
	case RunPending:
		r.runState = r.pendingState
	case RunPlanQueued:
		r.runState = r.planQueuedState
	case RunPlanning:
		r.runState = r.planningState
	case RunPlanned:
		r.runState = r.plannedState
	case RunPlannedAndFinished:
		r.runState = r.plannedAndFinishedState
	case RunApplyQueued:
		r.runState = r.applyQueuedState
	case RunApplying:
		r.runState = r.applyingState
	case RunApplied:
		r.runState = r.appliedState
	case RunDiscarded:
		r.runState = r.discardedState
	case RunErrored:
		r.runState = r.erroredState
	case RunCanceled:
		r.runState = r.canceledState
	default:
		return fmt.Errorf("no run state found corresponding to status %s", status)
	}
	return nil
}

// setupEnv invokes the necessary steps before a plan or apply can proceed.
func (r *Run) setupEnv(env Environment) error {
	if err := env.RunFunc(r.downloadConfig); err != nil {
		return err
	}
	err := env.RunFunc(func(ctx context.Context, env Environment) error {
		return deleteBackendConfigFromDirectory(ctx, env.Path())
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
	cv, err := env.ConfigurationVersionService().Download(ctx, r.configurationVersionID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}
	// Decompress and untar config
	if err := Unpack(bytes.NewBuffer(cv), env.Path()); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}
	return nil
}

// downloadState downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func (r *Run) downloadState(ctx context.Context, env Environment) error {
	state, err := env.StateVersionService().Current(ctx, r.workspaceID)
	if errors.Is(err, ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("retrieving current state version: %w", err)
	}
	statefile, err := env.StateVersionService().Download(ctx, state.ID())
	if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := os.WriteFile(filepath.Join(env.Path(), LocalStateFilename), statefile, 0644); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

func (r *Run) uploadPlan(ctx context.Context, env Environment) error {
	file, err := os.ReadFile(filepath.Join(env.Path(), PlanFilename))
	if err != nil {
		return err
	}

	if err := env.RunService().UploadPlanFile(ctx, r.Plan.ID(), file, PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Run) uploadJSONPlan(ctx context.Context, env Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.Path(), JSONPlanFilename))
	if err != nil {
		return err
	}
	if err := env.RunService().UploadPlanFile(ctx, r.Plan.ID(), jsonFile, PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

func (r *Run) downloadPlanFile(ctx context.Context, env Environment) error {
	plan, err := env.RunService().GetPlanFile(ctx, RunGetOptions{ID: String(r.ID())}, PlanFormatBinary)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(env.Path(), PlanFilename), plan, 0644)
}

// uploadState reads, parses, and uploads terraform state
func (r *Run) uploadState(ctx context.Context, env Environment) error {
	f, err := os.ReadFile(filepath.Join(env.Path(), LocalStateFilename))
	if err != nil {
		return err
	}
	state, err := UnmarshalState(f)
	if err != nil {
		return err
	}
	_, err = env.StateVersionService().Create(ctx, r.workspaceID, StateVersionCreateOptions{
		State:   String(base64.StdEncoding.EncodeToString(f)),
		MD5:     String(fmt.Sprintf("%x", md5.Sum(f))),
		Lineage: &state.Lineage,
		Serial:  Int64(state.Serial),
		Run:     r,
	})
	if err != nil {
		return err
	}
	return nil
}

type RunStatusTimestamp struct {
	Status    RunStatus
	Timestamp time.Time
}

// RunService implementations allow interactions with runs
type RunService interface {
	// Create a new run with the given options.
	Create(ctx context.Context, ws WorkspaceSpec, opts RunCreateOptions) (*Run, error)
	// Get retrieves a run with the given ID.
	Get(ctx context.Context, id string) (*Run, error)
	// List lists runs according to the given options.
	List(ctx context.Context, opts RunListOptions) (*RunList, error)
	// List and watch runs
	ListWatch(ctx context.Context, opts RunListOptions) (<-chan *Run, error)
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
	UploadPlanFile(ctx context.Context, planID string, plan []byte, format PlanFormat) error

	ReportService
}

// RunCreateOptions represents the options for creating a new run. See
// dto.RunCreateOptions for further detail.
type RunCreateOptions struct {
	IsDestroy              *bool
	Refresh                *bool
	RefreshOnly            *bool
	Message                *string
	ConfigurationVersionID *string
	WorkspaceID            *string
	TargetAddrs            []string
	ReplaceAddrs           []string
	WorkspaceSpec          WorkspaceSpec
}

// TestRunCreateOptions is for testing purposes only.
type TestRunCreateOptions struct {
	Speculative bool
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
	CreateRun(ctx context.Context, run *Run) error
	GetRun(ctx context.Context, opts RunGetOptions) (*Run, error)
	SetPlanFile(ctx context.Context, id string, file []byte, format PlanFormat) error
	GetPlanFile(ctx context.Context, id string, format PlanFormat) ([]byte, error)
	ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error)
	// UpdateStatus updates the run's status, providing a func with which to
	// perform updates in a transaction.
	UpdateStatus(ctx context.Context, opts RunGetOptions, fn func(*Run) error) (*Run, error)
	CreatePlanReport(ctx context.Context, planID string, report ResourceReport) error
	CreateApplyReport(ctx context.Context, applyID string, report ResourceReport) error
	DeleteRun(ctx context.Context, id string) error
}

// RunList represents a list of runs.
type RunList struct {
	*Pagination
	Items []*Run
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *RunList) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.RunList{
		Pagination: (*jsonapi.Pagination)(l.Pagination),
	}
	for _, item := range l.Items {
		dto.Items = append(dto.Items, item.ToJSONAPI(req).(*jsonapi.Run))
	}
	return dto
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
	// Get run via job ID
	JobID *string
	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include"`
}

func (o *RunGetOptions) String() string {
	if o.ID != nil {
		return *o.ID
	} else if o.PlanID != nil {
		return *o.PlanID
	} else if o.ApplyID != nil {
		return *o.ApplyID
	} else if o.JobID != nil {
		return *o.JobID
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
	// Filter by organization name
	OrganizationName *string `schema:"organization_name"`
	// Filter by workspace name
	WorkspaceName *string `schema:"workspace_name"`
}

// LogFields provides fields for logging
func (opts RunListOptions) LogFields() (fields []interface{}) {
	if opts.WorkspaceID != nil {
		fields = append(fields, "workspace_id", *opts.WorkspaceID)
	}
	if opts.WorkspaceName != nil {
		fields = append(fields, "name", *opts.WorkspaceName)
	}
	if opts.OrganizationName != nil {
		fields = append(fields, "organization", *opts.OrganizationName)
	}
	if len(opts.Statuses) > 0 {
		fields = append(fields, "status", fmt.Sprintf("%v", opts.Statuses))
	}
	return fields
}

func ContainsRunStatus(statuses []RunStatus, status RunStatus) bool {
	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}
