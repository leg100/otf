package run

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
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
	executionMode          otf.ExecutionMode
	positionInQueue        int
	refresh                bool
	refreshOnly            bool
	autoApply              bool
	speculative            bool
	status                 RunStatus
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
func (r *Run) RunID() string                          { return r.id }
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
func (r *Run) Status() RunStatus                      { return r.status }
func (r *Run) StatusTimestamps() []RunStatusTimestamp { return r.statusTimestamps }
func (r *Run) WorkspaceID() string                    { return r.workspaceID }
func (r *Run) ConfigurationVersionID() string         { return r.configurationVersionID }
func (r *Run) Plan() *Plan                            { return r.plan }
func (r *Run) Apply() *Apply                          { return r.apply }
func (r *Run) ExecutionMode() otf.ExecutionMode       { return r.executionMode }
func (r *Run) Commit() *string                        { return r.commit }

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
	_, err := r.Apply().StatusTimestamp(otf.PhaseRunning)
	return err == nil
}

// Phase returns the current phase.
func (r *Run) Phase() otf.PhaseType {
	switch r.status {
	case RunPending:
		return otf.PendingPhase
	case RunPlanQueued, RunPlanning, RunPlanned:
		return otf.PlanPhase
	case RunApplyQueued, RunApplying, RunApplied:
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
	r.updateStatus(RunDiscarded)

	if r.status == RunPending {
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
	case RunPending:
		r.plan.updateStatus(otf.PhaseUnreachable)
		r.apply.updateStatus(otf.PhaseUnreachable)
	case RunPlanQueued, RunPlanning:
		r.plan.updateStatus(otf.PhaseCanceled)
		r.apply.updateStatus(otf.PhaseUnreachable)
	case RunApplyQueued, RunApplying:
		r.apply.updateStatus(otf.PhaseCanceled)
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
func (r *Run) Do(env otf.Environment) error {
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
func (r *Run) EnqueuePlan() error {
	if r.status != RunPending {
		return fmt.Errorf("cannot enqueue run with status %s", r.status)
	}
	r.updateStatus(RunPlanQueued)
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
	if r.status != RunPlanned {
		return fmt.Errorf("cannot apply run")
	}
	r.updateStatus(RunApplyQueued)
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
	case RunPlanQueued:
		return r.startPlan()
	case RunApplyQueued:
		return r.startApply()
	case RunPlanning, RunApplying:
		return otf.ErrPhaseAlreadyStarted
	default:
		return ErrInvalidRunStateTransition
	}
}

// Finish updates the run to reflect its plan or apply phase having finished.
func (r *Run) Finish(phase otf.PhaseType, opts otf.PhaseFinishOptions) error {
	if r.status == RunCanceled {
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

func (r *Run) startPlan() error {
	if r.status != RunPlanQueued {
		return ErrInvalidRunStateTransition
	}
	r.updateStatus(RunPlanning)
	r.plan.updateStatus(otf.PhaseRunning)
	return nil
}

func (r *Run) startApply() error {
	if r.status != RunApplyQueued {
		return ErrInvalidRunStateTransition
	}
	r.updateStatus(RunApplying)
	r.apply.updateStatus(otf.PhaseRunning)
	return nil
}

func (r *Run) finishPlan(opts otf.PhaseFinishOptions) error {
	if r.status != RunPlanning {
		return ErrInvalidRunStateTransition
	}
	if opts.Errored {
		r.updateStatus(RunErrored)
		r.plan.updateStatus(otf.PhaseErrored)
		r.apply.updateStatus(otf.PhaseUnreachable)
		return nil
	}

	r.updateStatus(RunPlanned)
	r.plan.updateStatus(otf.PhaseFinished)

	if !r.HasChanges() || r.Speculative() {
		r.updateStatus(RunPlannedAndFinished)
		r.apply.updateStatus(otf.PhaseUnreachable)
	} else if r.autoApply {
		return r.EnqueueApply()
	}
	return nil
}

func (r *Run) finishApply(opts otf.PhaseFinishOptions) error {
	if r.status != RunApplying {
		return ErrInvalidRunStateTransition
	}
	if opts.Errored {
		r.updateStatus(RunErrored)
		r.apply.updateStatus(otf.PhaseErrored)
	} else {
		r.updateStatus(RunApplied)
		r.apply.updateStatus(otf.PhaseFinished)
	}
	return nil
}

func (r *Run) updateStatus(status RunStatus) {
	r.status = status
	r.statusTimestamps = append(r.statusTimestamps, RunStatusTimestamp{
		Status:    status,
		Timestamp: otf.CurrentTimestamp(),
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

func (r *Run) doPlan(env otf.Environment) error {
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

func (r *Run) doApply(env otf.Environment) error {
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
func (r *Run) doTerraformPlan(env otf.Environment) error {
	var args []string
	if r.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+PlanFilename)
	return env.RunTerraform("plan", args...)
}

// doTerraformApply invokes terraform apply
func (r *Run) doTerraformApply(env otf.Environment) error {
	var args []string
	if r.isDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, PlanFilename)
	return env.RunTerraform("apply", args...)
}

// setupEnv invokes the necessary steps before a plan or apply can proceed.
func (r *Run) setupEnv(env otf.Environment) error {
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
		// Download lock file from plan phase for the apply phase, to ensure
		// same providers are used in both phases.
		if err := env.RunFunc(r.downloadLockFile); err != nil {
			return err
		}
	}
	if err := env.RunTerraform("init"); err != nil {
		return fmt.Errorf("running terraform init: %w", err)
	}
	return nil
}

func (r *Run) deleteBackendConfig(ctx context.Context, env otf.Environment) error {
	if err := otf.RewriteHCL(env.WorkingDir(), otf.RemoveBackendBlock); err != nil {
		return fmt.Errorf("removing backend config: %w", err)
	}
	return nil
}

func (r *Run) downloadTerraform(ctx context.Context, env otf.Environment) error {
	ws, err := env.GetWorkspace(ctx, r.workspaceID)
	if err != nil {
		return err
	}
	_, err = env.Download(ctx, ws.TerraformVersion(), env)
	if err != nil {
		return err
	}
	return nil
}

func (r *Run) downloadConfig(ctx context.Context, env otf.Environment) error {
	// Download config
	cv, err := env.DownloadConfig(ctx, r.configurationVersionID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}
	// Decompress and untar config into environment root
	if err := otf.Unpack(bytes.NewBuffer(cv), env.Path()); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}
	return nil
}

// downloadState downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func (r *Run) downloadState(ctx context.Context, env otf.Environment) error {
	statefile, err := env.DownloadCurrentState(ctx, r.workspaceID)
	if errors.Is(err, otf.ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := os.WriteFile(filepath.Join(env.WorkingDir(), LocalStateFilename), statefile, 0o644); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

func (r *Run) uploadPlan(ctx context.Context, env otf.Environment) error {
	file, err := os.ReadFile(filepath.Join(env.WorkingDir(), PlanFilename))
	if err != nil {
		return err
	}

	if err := env.UploadPlanFile(ctx, r.id, file, PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Run) uploadJSONPlan(ctx context.Context, env otf.Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.WorkingDir(), JSONPlanFilename))
	if err != nil {
		return err
	}
	if err := env.UploadPlanFile(ctx, r.id, jsonFile, PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

// downloadLockFile downloads the .terraform.lock.hcl file into the working
// directory. If one has not been uploaded then this will simply write an empty
// file, which is harmless.
func (r *Run) downloadLockFile(ctx context.Context, env otf.Environment) error {
	lockFile, err := env.GetLockFile(ctx, r.id)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(env.WorkingDir(), LockFilename), lockFile, 0o644)
}

func (r *Run) uploadLockFile(ctx context.Context, env otf.Environment) error {
	lockFile, err := os.ReadFile(filepath.Join(env.WorkingDir(), LockFilename))
	if errors.Is(err, fs.ErrNotExist) {
		// there is no lock file to upload, which is ok
		return nil
	} else if err != nil {
		return errors.Wrap(err, "reading lock file")
	}
	if err := env.UploadLockFile(ctx, r.id, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (r *Run) downloadPlanFile(ctx context.Context, env otf.Environment) error {
	plan, err := env.GetPlanFile(ctx, r.id, PlanFormatBinary)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(env.WorkingDir(), PlanFilename), plan, 0o644)
}

// uploadState reads, parses, and uploads terraform state
func (r *Run) uploadState(ctx context.Context, env otf.Environment) error {
	state, err := os.ReadFile(filepath.Join(env.WorkingDir(), LocalStateFilename))
	if err != nil {
		return err
	}
	err = env.CreateStateVersion(ctx, otf.CreateStateVersionOptions{
		WorkspaceID: &r.workspaceID,
		State:       state,
	})
	return err
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
