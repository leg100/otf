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
	executionMode          ExecutionMode
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
	latest                 bool

	commit *string // commit sha that triggered this run

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
func (r *Run) Organization() string                   { return r.organizationName }
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
func (r *Run) ExecutionMode() ExecutionMode           { return r.executionMode }

func (r *Run) Commit() *string { return r.commit }

// Latest determines whether run is the latest run for a workspace, i.e.
// its current run, or the most recent current run.
func (r *Run) Latest() bool { return r.latest }

func (r *Run) Queued() bool {
	return r.status == RunPlanQueued || r.status == RunApplyQueued
}

func (r *Run) HasChanges() bool {
	return r.plan.HasChanges()
}

func (r *Run) PlanOnly() bool {
	return r.status == RunPlannedAndFinished
}

// HasApply determines whether the run has started applying yet.
func (r *Run) HasApply() bool {
	_, err := r.Apply().StatusTimestamp(PhaseRunning)
	return err == nil
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
	if !r.Discardable() {
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
	if !r.Cancelable() {
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

// EnqueuePlan enqueues a plan for the run. It also sets the run as the latest
// run for its workspace (speculative runs are ignored).
func (r *Run) EnqueuePlan(ctx context.Context, svc WorkspaceLockService) error {
	if r.status != RunPending {
		return fmt.Errorf("cannot enqueue run with status %s", r.status)
	}
	r.updateStatus(RunPlanQueued)
	r.plan.updateStatus(PhaseQueued)

	if !r.Speculative() {
		// Lock the workspace on behalf of the run
		ctx = AddSubjectToContext(ctx, r)
		_, err := svc.LockWorkspace(ctx, WorkspaceSpec{ID: String(r.WorkspaceID())}, WorkspaceLockOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (*Run) CanAccessSite(action Action) bool {
	// run cannot carry out site-level actions
	return false
}

func (r *Run) CanAccessOrganization(action Action, name string) bool {
	// run cannot access organization-level resources
	return false
}

func (r *Run) CanAccessWorkspace(action Action, policy *WorkspacePolicy) bool {
	// run can access anything within its workspace
	return r.workspaceID == policy.WorkspaceID
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
}

// IncludeWorkspace adds a workspace for inclusion in the run's JSON-API object.
//
// TODO: remove; instead retrieve JSON-API inclusions in the http pkg
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
		return nil
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

// Discardable determines whether run can be discarded.
func (r *Run) Discardable() bool {
	switch r.Status() {
	case RunPending, RunPlanned:
		return true
	default:
		return false
	}
}

// Cancelable determines whether run can be cancelled.
func (r *Run) Cancelable() bool {
	switch r.Status() {
	case RunPending, RunPlanQueued, RunPlanning, RunPlanned, RunApplyQueued, RunApplying:
		return true
	default:
		return false
	}
}

// Confirmable determines whether run can be confirmed.
func (r *Run) Confirmable() bool {
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
	if err := env.RunCLI("sh", "-c", fmt.Sprintf("%s show -json %s > %s", env.TerraformPath(), PlanFilename, JSONPlanFilename)); err != nil {
		return err
	}
	if err := env.RunFunc(r.uploadPlan); err != nil {
		return err
	}
	if err := env.RunFunc(r.uploadJSONPlan); err != nil {
		return err
	}
	// upload lock file for use in the apply phase - see note in setupEnv.
	if err := env.RunFunc(r.uploadLockFile); err != nil {
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
	var args []string
	if r.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+PlanFilename)
	return env.RunTerraform("plan", args...)
}

// doTerraformApply invokes terraform apply
func (r *Run) doTerraformApply(env Environment) error {
	var args []string
	if r.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, PlanFilename)
	return env.RunTerraform("apply", args...)
}

// setupEnv invokes the necessary steps before a plan or apply can proceed.
func (r *Run) setupEnv(env Environment) error {
	if err := env.RunFunc(r.downloadTerraform); err != nil {
		return err
	}
	if err := env.RunFunc(r.downloadConfig); err != nil {
		return err
	}
	if err := env.RunFunc(r.deleteBackendConfig); err != nil {
		return err
	}
	if err := env.RunFunc(r.downloadState); err != nil {
		return err
	}
	if r.status == RunApplying {
		// Download lock file in apply phase - the user *should* have pushed
		// their own lock file and otf then treats it as immutable and respects
		// the provider versions it specifies. However, to cover the instances
		// in which they don't push a lock file, then otf generates the lock
		// file in the plan phase and the same file is persisted here to ensure
		// the exact same providers are used in both phases.
		if err := env.RunFunc(r.downloadLockFile); err != nil {
			return err
		}
	}
	if err := env.RunTerraform("init"); err != nil {
		return fmt.Errorf("running terraform init: %w", err)
	}
	return nil
}

func (r *Run) deleteBackendConfig(ctx context.Context, env Environment) error {
	if err := rewriteHCL(env.Path(), removeBackendBlock); err != nil {
		return fmt.Errorf("removing backend config: %w", err)
	}
	return nil
}

func (r *Run) downloadTerraform(ctx context.Context, env Environment) error {
	ws, err := env.GetWorkspace(ctx, WorkspaceSpec{ID: &r.workspaceID})
	if err != nil {
		return err
	}
	_, err = env.Download(ctx, ws.TerraformVersion(), env)
	if err != nil {
		return err
	}
	return nil
}

func (r *Run) downloadConfig(ctx context.Context, env Environment) error {
	// Download config
	cv, err := env.DownloadConfig(ctx, r.configurationVersionID)
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
	state, err := env.CurrentStateVersion(ctx, r.workspaceID)
	if errors.Is(err, ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("retrieving current state version: %w", err)
	}
	statefile, err := env.DownloadState(ctx, state.ID())
	if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := os.WriteFile(filepath.Join(env.Path(), LocalStateFilename), statefile, 0o644); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

func (r *Run) uploadPlan(ctx context.Context, env Environment) error {
	file, err := os.ReadFile(filepath.Join(env.Path(), PlanFilename))
	if err != nil {
		return err
	}

	if err := env.UploadPlanFile(ctx, r.id, file, PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Run) uploadJSONPlan(ctx context.Context, env Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.Path(), JSONPlanFilename))
	if err != nil {
		return err
	}
	if err := env.UploadPlanFile(ctx, r.id, jsonFile, PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

func (r *Run) downloadLockFile(ctx context.Context, env Environment) error {
	lockFile, err := env.GetLockFile(ctx, r.id)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(env.Path(), LockFilename), lockFile, 0o644)
}

func (r *Run) uploadLockFile(ctx context.Context, env Environment) error {
	lockFile, err := os.ReadFile(filepath.Join(env.Path(), LockFilename))
	if err != nil {
		return err
	}
	if err := env.UploadLockFile(ctx, r.id, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (r *Run) downloadPlanFile(ctx context.Context, env Environment) error {
	plan, err := env.GetPlanFile(ctx, r.id, PlanFormatBinary)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(env.Path(), PlanFilename), plan, 0o644)
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
	_, err = env.CreateStateVersion(ctx, r.workspaceID, StateVersionCreateOptions{
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
	CreateRun(ctx context.Context, ws WorkspaceSpec, opts RunCreateOptions) (*Run, error)
	// Get retrieves a run with the given ID.
	GetRun(ctx context.Context, id string) (*Run, error)
	// List lists runs according to the given options.
	ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error)
	// Delete deletes a run with the given ID.
	DeleteRun(ctx context.Context, id string) error
	// EnqueuePlan enqueues a plan
	EnqueuePlan(ctx context.Context, id string) (*Run, error)
	// Apply a run with the given ID.
	//
	// TODO: return run
	ApplyRun(ctx context.Context, id string, opts RunApplyOptions) error
	// Discard discards a run with the given ID.
	//
	// TODO: return run
	DiscardRun(ctx context.Context, id string, opts RunDiscardOptions) error
	// Cancel run.
	//
	// TODO: return run
	CancelRun(ctx context.Context, id string, opts RunCancelOptions) error
	// Forcefully cancel a run.
	//
	// TODO: return run
	ForceCancelRun(ctx context.Context, id string, opts RunForceCancelOptions) error
	// Start a run phase.
	StartPhase(ctx context.Context, id string, phase PhaseType, opts PhaseStartOptions) (*Run, error)
	// Finish a run phase.
	FinishPhase(ctx context.Context, id string, phase PhaseType, opts PhaseFinishOptions) (*Run, error)
	// GetPlanFile retrieves a run's plan file with the requested format.
	GetPlanFile(ctx context.Context, id string, format PlanFormat) ([]byte, error)
	// UploadPlanFile saves a run's plan file with the requested format.
	UploadPlanFile(ctx context.Context, id string, plan []byte, format PlanFormat) error
	// GetLockFile retrieves a run's lock file (.terraform.lock.hcl)
	GetLockFile(ctx context.Context, id string) ([]byte, error)
	// UploadLockFile saves a run's lock file (.terraform.lock.hcl)
	UploadLockFile(ctx context.Context, id string, lockFile []byte) error
	// Read and write logs for run phases.
	LogService
	// Tail logs of a run phase
	Tail(ctx context.Context, opts GetChunkOptions) (<-chan Chunk, error)
	// StartRun creates and starts a run.
	StartRun(ctx context.Context, spec WorkspaceSpec, opts ConfigurationVersionCreateOptions) (*Run, error)
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

// RunApplyOptions represents the options for applying a run.
type RunApplyOptions struct {
	// An optional comment about the run.
	Comment *string `json:"comment,omitempty"`
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
	SetLockFile(ctx context.Context, id string, file []byte) error
	GetLockFile(ctx context.Context, id string) ([]byte, error)
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

// RunListOptions are options for paginating and filtering a list of runs
type RunListOptions struct {
	ListOptions
	// Filter by run statuses (with an implicit OR condition)
	Statuses []RunStatus `schema:"statuses,omitempty"`
	// Filter by workspace ID
	WorkspaceID *string `schema:"workspace_id,omitempty"`
	// Filter by organization name
	OrganizationName *string `schema:"organization_name,omitempty"`
	// Filter by workspace name
	WorkspaceName *string `schema:"workspace_name,omitempty"`
	// Filter by speculative or non-speculative
	Speculative *bool `schema:"-"`
	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include,omitempty"`
}

// LogFields provides fields for logging
//
// TODO: use logr marshaller instead
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
