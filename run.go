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
	ErrStatusTimestampNotFound   = errors.New("corresponding status timestamp not found")
	ErrRunDiscardNotAllowed      = errors.New("run was not paused for confirmation or priority; discard not allowed")
	ErrRunCancelNotAllowed       = errors.New("run was not planning or applying; cancel not allowed")
	ErrRunForceCancelNotAllowed  = errors.New("run was not planning or applying, has not been canceled non-forcefully, or the cool-off period has not yet passed")
	ErrInvalidRunGetOptions      = errors.New("invalid run get options")
	ErrInvalidRunStateTransition = errors.New("invalid run state transition")
	ActiveRun                    = []RunStatus{
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

// RunStatus represents a run state.
type RunStatus string

func (r RunStatus) String() string { return string(r) }

type Run struct {
	id                     string
	createdAt              time.Time
	forceCancelAvailableAt *time.Time
	isDestroy              bool
	message                string
	positionInQueue        int
	refresh                bool
	refreshOnly            bool
	autoApply              bool
	speculative            bool
	status                 RunStatus
	statusTimestamps       []RunStatusTimestamp
	replaceAddrs           []string
	targetAddrs            []string
	organizationName       string
	workspaceName          string
	workspaceID            string
	configurationVersionID string
	// Relations
	plan      *Plan
	apply     *Apply
	workspace *Workspace
}

func (r *Run) ID() string                             { return r.id }
func (r *Run) RunID() string                          { return r.id }
func (r *Run) CreatedAt() time.Time                   { return r.createdAt }
func (r *Run) String() string                         { return r.id }
func (r *Run) IsDestroy() bool                        { return r.isDestroy }
func (r *Run) ForceCancelAvailableAt() *time.Time     { return r.forceCancelAvailableAt }
func (r *Run) Message() string                        { return r.message }
func (r *Run) OrganizationName() string               { return r.organizationName }
func (r *Run) Refresh() bool                          { return r.refresh }
func (r *Run) RefreshOnly() bool                      { return r.refreshOnly }
func (r *Run) ReplaceAddrs() []string                 { return r.replaceAddrs }
func (r *Run) TargetAddrs() []string                  { return r.targetAddrs }
func (r *Run) Speculative() bool                      { return r.speculative }
func (r *Run) Status() RunStatus                      { return r.status }
func (r *Run) StatusTimestamps() []RunStatusTimestamp { return r.statusTimestamps }
func (r *Run) WorkspaceName() string                  { return r.workspaceName }
func (r *Run) WorkspaceID() string                    { return r.workspaceID }
func (r *Run) Workspace() *Workspace                  { return r.workspace }
func (r *Run) ConfigurationVersionID() string         { return r.configurationVersionID }
func (r *Run) Plan() *Plan                            { return r.plan }
func (r *Run) Apply() *Apply                          { return r.apply }

func (r *Run) Queued() bool {
	return r.status == RunPlanQueued || r.status == RunApplyQueued
}

func (r *Run) HasChanges() bool {
	return r.plan.HasChanges()
}

func (r *Run) PlanOnly() bool {
	return r.status == RunPlannedAndFinished
}

// Phase returns the current phase.
func (r *Run) Phase() PhaseType {
	switch r.status {
	case RunPending:
		return PendingPhase
	case RunPlanQueued, RunPlanning, RunPlanned:
		return PlanPhase
	case RunApplyQueued, RunApplying, RunApplied:
		return ApplyPhase
	default:
		return UnknownPhase
	}
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.discardable() {
		return ErrRunDiscardNotAllowed
	}
	r.updateStatus(RunDiscarded)

	if r.status == RunPending {
		r.plan.updateStatus(PhaseUnreachable)
	}
	r.apply.updateStatus(PhaseUnreachable)

	return nil
}

// Cancel run. Returns a boolean indicating whether a cancel request should be
// enqueued (for an agent to kill an in progress process)
func (r *Run) Cancel() (enqueue bool, err error) {
	if !r.cancelable() {
		return false, ErrRunCancelNotAllowed
	}
	// permit run to be force canceled after a cool off period of 10 seconds has
	// elapsed.
	tenSecondsFromNow := CurrentTimestamp().Add(10 * time.Second)
	r.forceCancelAvailableAt = &tenSecondsFromNow

	switch r.status {
	case RunPending:
		r.plan.updateStatus(PhaseUnreachable)
		r.apply.updateStatus(PhaseUnreachable)
	case RunPlanQueued, RunPlanning:
		r.plan.updateStatus(PhaseCanceled)
		r.apply.updateStatus(PhaseUnreachable)
	case RunApplyQueued, RunApplying:
		r.apply.updateStatus(PhaseCanceled)
	}

	if r.status == RunPlanning || r.status == RunApplying {
		enqueue = true
	}

	r.updateStatus(RunCanceled)

	return enqueue, nil
}

// ForceCancel force cancels a run. A cool-off period of 10 seconds must have
// elapsed following a cancelation request before a run can be force canceled.
func (r *Run) ForceCancel() error {
	if r.forceCancelAvailableAt != nil && time.Now().After(*r.forceCancelAvailableAt) {
		r.updateStatus(RunCanceled)
		return nil
	}
	return ErrRunForceCancelNotAllowed
}

// Do executes the current phase
func (r *Run) Do(env Environment) error {
	if err := r.setupEnv(env); err != nil {
		return err
	}
	switch r.status {
	case RunPlanning:
		return r.doPlan(env)
	case RunApplying:
		return r.doApply(env)
	default:
		return fmt.Errorf("invalid status: %s", r.status)
	}
}

// Done determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (r *Run) Done() bool {
	switch r.Status() {
	case RunApplied, RunPlannedAndFinished, RunDiscarded, RunCanceled, RunErrored:
		return true
	default:
		return false
	}
}

func (r *Run) EnqueuePlan() error {
	if r.status != RunPending {
		return fmt.Errorf("cannot enqueue pending run")
	}
	r.updateStatus(RunPlanQueued)
	r.plan.updateStatus(PhaseQueued)
	return nil
}

func (r *Run) EnqueueApply() error {
	if r.status != RunPlanned {
		return fmt.Errorf("cannot apply run")
	}
	r.updateStatus(RunApplyQueued)
	r.apply.updateStatus(PhaseQueued)
	return nil
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

// Start a run phase
func (r *Run) Start(phase PhaseType) error {
	switch r.status {
	case RunPlanQueued:
		return r.startPlan()
	case RunApplyQueued:
		return r.startApply()
	case RunPlanning, RunApplying:
		return ErrPhaseAlreadyStarted
	default:
		return ErrInvalidRunStateTransition
	}
	return nil
}

// Finish updates the run to reflect its plan or apply phase having finished.
func (r *Run) Finish(phase PhaseType, opts PhaseFinishOptions) error {
	if r.status == RunCanceled {
		// run was canceled before the phase finished so nothing more to do.
		return nil
	}
	switch phase {
	case PlanPhase:
		return r.finishPlan(opts)
	case ApplyPhase:
		return r.finishApply(opts)
	default:
		return fmt.Errorf("unknown phase")
	}
	return nil
}

// ToJSONAPI assembles a JSON-API DTO.
func (r *Run) ToJSONAPI(req *http.Request) any {
	dto := &jsonapi.Run{
		ID: r.ID(),
		Actions: &jsonapi.RunActions{
			IsCancelable:      r.cancelable(),
			IsConfirmable:     r.confirmable(),
			IsForceCancelable: r.forceCancelAvailableAt != nil,
			IsDiscardable:     r.discardable(),
		},
		CreatedAt:              r.CreatedAt(),
		ForceCancelAvailableAt: r.forceCancelAvailableAt,
		HasChanges:             r.plan.HasChanges(),
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
		Apply: r.apply.ToJSONAPI(req).(*jsonapi.Apply),
		Plan:  r.plan.ToJSONAPI(req).(*jsonapi.Plan),
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

// IncludeWorkspace adds a workspace for inclusion in the run's JSON-API object.
func (r *Run) IncludeWorkspace(ws *Workspace) {
	r.workspace = ws
}

func (r *Run) startPlan() error {
	if r.status != RunPlanQueued {
		return ErrInvalidRunStateTransition
	}
	r.updateStatus(RunPlanning)
	r.plan.updateStatus(PhaseRunning)
	return nil
}

func (r *Run) startApply() error {
	if r.status != RunApplyQueued {
		return ErrInvalidRunStateTransition
	}
	r.updateStatus(RunApplying)
	r.apply.updateStatus(PhaseRunning)
	return nil
}

func (r *Run) finishPlan(opts PhaseFinishOptions) error {
	if r.status != RunPlanning {
		return ErrInvalidRunStateTransition
	}
	if opts.Errored {
		r.updateStatus(RunErrored)
		r.plan.updateStatus(PhaseErrored)
		r.apply.updateStatus(PhaseUnreachable)
	}

	r.updateStatus(RunPlanned)
	r.plan.updateStatus(PhaseFinished)

	if !r.HasChanges() || r.Speculative() {
		r.updateStatus(RunPlannedAndFinished)
		r.apply.updateStatus(PhaseUnreachable)
	} else if r.autoApply {
		return r.EnqueueApply()
	}
	return nil
}

func (r *Run) finishApply(opts PhaseFinishOptions) error {
	if r.status != RunApplying {
		return ErrInvalidRunStateTransition
	}
	if opts.Errored {
		r.updateStatus(RunErrored)
		r.apply.updateStatus(PhaseErrored)
	} else {
		r.updateStatus(RunApplied)
		r.apply.updateStatus(PhaseFinished)
	}
	return nil
}

func (r *Run) updateStatus(status RunStatus) {
	r.status = status
	r.statusTimestamps = append(r.statusTimestamps, RunStatusTimestamp{
		Status:    status,
		Timestamp: CurrentTimestamp(),
	})
}

// discardable determines whether run can be discarded.
func (r *Run) discardable() bool {
	switch r.Status() {
	case RunPending, RunPlanned:
		return true
	default:
		return false
	}
}

// cancelable determines whether run can be cancelled.
func (r *Run) cancelable() bool {
	switch r.Status() {
	case RunPending, RunPlanQueued, RunPlanning, RunApplyQueued, RunApplying:
		return true
	default:
		return false
	}
}

// confirmable determines whether run can be confirmed.
func (r *Run) confirmable() bool {
	switch r.Status() {
	case RunPlanned:
		return true
	default:
		return false
	}
}

func (r *Run) doPlan(env Environment) error {
	if err := r.doTerraformPlan(env); err != nil {
		return err
	}
	if err := env.RunCLI("sh", "-c", fmt.Sprintf("terraform show -json %s > %s", PlanFilename, JSONPlanFilename)); err != nil {
		return err
	}
	if err := env.RunFunc(r.uploadPlan); err != nil {
		return err
	}
	if err := env.RunFunc(r.uploadJSONPlan); err != nil {
		return err
	}
	return nil
}

func (r *Run) doApply(env Environment) error {
	if err := env.RunFunc(r.downloadPlanFile); err != nil {
		return err
	}
	if err := r.doTerraformApply(env); err != nil {
		return err
	}
	if err := env.RunFunc(r.uploadState); err != nil {
		return err
	}
	return nil
}

// doTerraformPlan invokes terraform plan
func (r *Run) doTerraformPlan(env Environment) error {
	args := []string{
		"plan",
	}
	if r.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+PlanFilename)
	return env.RunCLI("terraform", args...)
}

// doTerraformApply invokes terraform apply
func (r *Run) doTerraformApply(env Environment) error {
	args := []string{"apply"}
	if r.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, PlanFilename)
	return env.RunCLI("terraform", args...)
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

	if err := env.RunService().UploadPlanFile(ctx, r.id, file, PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Run) uploadJSONPlan(ctx context.Context, env Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.Path(), JSONPlanFilename))
	if err != nil {
		return err
	}
	if err := env.RunService().UploadPlanFile(ctx, r.id, jsonFile, PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

func (r *Run) downloadPlanFile(ctx context.Context, env Environment) error {
	plan, err := env.RunService().GetPlanFile(ctx, r.id, PlanFormatBinary)
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
	// EnqueuePlan enqueues a plan
	EnqueuePlan(ctx context.Context, id string) (*Run, error)
	// Apply a run with the given ID.
	Apply(ctx context.Context, id string, opts RunApplyOptions) error
	// Discard discards a run with the given ID.
	Discard(ctx context.Context, id string, opts RunDiscardOptions) error
	// Cancel run.
	Cancel(ctx context.Context, id string, opts RunCancelOptions) error
	// Forcefully cancel a run.
	ForceCancel(ctx context.Context, id string, opts RunForceCancelOptions) error
	// Start a run phase.
	Start(ctx context.Context, id string, phase PhaseType, opts PhaseStartOptions) (*Run, error)
	// Finish a run phase.
	Finish(ctx context.Context, id string, phase PhaseType, opts PhaseFinishOptions) (*Run, error)
	// GetPlanFile retrieves a run's plan file with the requested format.
	GetPlanFile(ctx context.Context, id string, format PlanFormat) ([]byte, error)
	// UploadPlanFile saves a run's plan file with the requested format.
	UploadPlanFile(ctx context.Context, id string, plan []byte, format PlanFormat) error
	// Read and write logs for run phases.
	LogService
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
	Status      RunStatus
	AutoApply   bool
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
	GetRun(ctx context.Context, id string) (*Run, error)
	SetPlanFile(ctx context.Context, id string, file []byte, format PlanFormat) error
	GetPlanFile(ctx context.Context, id string, format PlanFormat) ([]byte, error)
	ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error)
	// UpdateStatus updates the run's status, providing a func with which to
	// perform updates in a transaction.
	UpdateStatus(ctx context.Context, id string, fn func(*Run) error) (*Run, error)
	CreatePlanReport(ctx context.Context, id string, report ResourceReport) error
	CreateApplyReport(ctx context.Context, id string, report ResourceReport) error
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

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	ListOptions
	// Order: oldest first or newest first
	Order ListOrder
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
