// Package run is responsible for OTF runs, the primary mechanism for executing
// terraform
package run

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

const (
	PlanFormatBinary = "bin"  // plan file in binary format
	PlanFormatJSON   = "json" // plan file in json format

	PlanOnlyOperation     Operation = "plan-only"
	PlanAndApplyOperation Operation = "plan-and-apply"
	DestroyAllOperation   Operation = "destroy-all"

	// defaultRefresh specifies that the state be refreshed prior to running a
	// plan
	defaultRefresh = true
)

var ErrInvalidRunStateTransition = errors.New("invalid run state transition")

type (
	PlanFormat string

	// Run operation specifies the terraform execution mode.
	Operation string

	// Run is a terraform run.
	Run struct {
		ID                     resource.TfeID          `jsonapi:"primary,runs"`
		CreatedAt              time.Time               `jsonapi:"attribute" json:"created_at"`
		UpdatedAt              time.Time               `jsonapi:"attribute" json:"updated_at"`
		IsDestroy              bool                    `jsonapi:"attribute" json:"is_destroy"`
		CancelSignaledAt       *time.Time              `jsonapi:"attribute" json:"cancel_signaled_at"`
		Message                string                  `jsonapi:"attribute" json:"message"`
		Organization           organization.Name       `jsonapi:"attribute" json:"organization"`
		Refresh                bool                    `jsonapi:"attribute" json:"refresh"`
		RefreshOnly            bool                    `jsonapi:"attribute" json:"refresh_only"`
		ReplaceAddrs           []string                `jsonapi:"attribute" json:"replace_addrs"`
		PositionInQueue        int                     `jsonapi:"attribute" json:"position_in_queue"`
		TargetAddrs            []string                `jsonapi:"attribute" json:"target_addrs"`
		EngineVersion          string                  `jsonapi:"attribute" json:"engine_version"`
		Engine                 *engine.Engine          `jsonapi:"attribute" json:"engine"`
		AllowEmptyApply        bool                    `jsonapi:"attribute" json:"allow_empty_apply"`
		AutoApply              bool                    `jsonapi:"attribute" json:"auto_apply"`
		PlanOnly               bool                    `jsonapi:"attribute" json:"plan_only"`
		Source                 source.Source           `jsonapi:"attribute" json:"source"`
		Status                 runstatus.Status        `jsonapi:"attribute" json:"status"`
		WorkspaceID            resource.TfeID          `jsonapi:"attribute" json:"workspace_id"`
		ConfigurationVersionID resource.TfeID          `jsonapi:"attribute" json:"configuration_version_id"`
		ExecutionMode          workspace.ExecutionMode `jsonapi:"attribute" json:"execution_mode"`
		Variables              []Variable              `jsonapi:"attribute" json:"variables"`
		Plan                   Phase                   `jsonapi:"attribute" json:"plan"`
		Apply                  Phase                   `jsonapi:"attribute" json:"apply"`

		// Timestamps of when a state transition occured. Ordered earliest
		// first.
		StatusTimestamps []StatusTimestamp `jsonapi:"attribute" json:"status_timestamps"`

		Latest bool `jsonapi:"attribute" json:"latest"` // is latest run for workspace

		// IngressAttributes is non-nil if run was triggered by a VCS event.
		IngressAttributes *configversion.IngressAttributes `jsonapi:"attribute" json:"ingress_attributes"`

		// Username of user who created the run. This is nil if the run was
		// instead triggered by a VCS event.
		CreatedBy *user.Username

		// OTF doesn't support cost estimation but some go-tfe API tests expect
		// a run to enter the RunCostEstimated state, and this boolean
		// determines whether to enter that state upon finishing a plan.
		CostEstimationEnabled bool
	}

	Variable struct {
		Key   string
		Value string
	}

	StatusTimestamp struct {
		Status    runstatus.Status `json:"status"`
		Timestamp time.Time        `json:"timestamp"`
	}

	// CreateOptions represents the options for creating a new run. See
	// api.types.RunCreateOptions for documentation on each field.
	CreateOptions struct {
		IsDestroy   *bool
		Refresh     *bool
		RefreshOnly *bool
		Message     *string
		// Specifies the configuration version to use for this run. If the
		// configuration version ID is nil, the run will be created using the
		// workspace's latest configuration version.
		ConfigurationVersionID *resource.TfeID
		TargetAddrs            []string
		ReplaceAddrs           []string
		AutoApply              *bool
		Source                 source.Source
		TerraformVersion       *string
		AllowEmptyApply        *bool
		// PlanOnly specifies if this is a speculative, plan-only run that
		// Terraform cannot apply. Takes precedence over whether the
		// configuration version is marked as speculative or not.
		PlanOnly      *bool
		Variables     []Variable
		CreatedBy     *user.Username
		EngineVersion string

		costEstimationEnabled bool

		// testing purposes
		now *time.Time
	}

	// ListOptions are options for paginating and filtering a list of runs
	ListOptions struct {
		resource.PageOptions
		// Filter by workspace ID
		WorkspaceID *resource.TfeID `schema:"workspace_id,omitempty"`
		// Filter by organization name
		Organization *organization.Name `schema:"organization_name,omitempty"`
		// Filter by workspace name
		WorkspaceName *string `schema:"workspace_name,omitempty"`
		// Filter by run statuses (with an implicit OR condition)
		Statuses []runstatus.Status `schema:"search[status],omitempty"`
		// Filter by plan-only runs
		PlanOnly *bool `schema:"-"`
		// Filter by sources
		Sources []source.Source
		// Filter by commit SHA that triggered a run
		CommitSHA *string
		// Filter by VCS user's username that triggered a run
		VCSUsername *string
		// Filter by run's time of creation - list only runs that were created
		// before this date.
		BeforeCreatedAt *time.Time
	}

	// WatchOptions filters events returned by the Watch endpoint.
	WatchOptions struct {
		Organization *organization.Name `schema:"organization_name,omitempty"` // filter by organization name
		WorkspaceID  *resource.TfeID    `schema:"workspace_id,omitempty"`      // filter by workspace ID; mutually exclusive with organization filter
	}
)

// NewRun constructs a new run. NOTE: you probably want to use factory.NewRun
// instead.
func NewRun(
	ws *workspace.Workspace,
	cv *configversion.ConfigurationVersion,
	opts CreateOptions,
) (*Run, error) {
	run := Run{
		ID:                     resource.NewTfeID(resource.RunKind),
		CreatedAt:              internal.CurrentTimestamp(opts.now),
		Refresh:                defaultRefresh,
		Organization:           ws.Organization,
		ConfigurationVersionID: cv.ID,
		WorkspaceID:            ws.ID,
		PlanOnly:               cv.Speculative,
		ReplaceAddrs:           opts.ReplaceAddrs,
		TargetAddrs:            opts.TargetAddrs,
		ExecutionMode:          ws.ExecutionMode,
		AutoApply:              ws.AutoApply,
		IngressAttributes:      cv.IngressAttributes,
		Source:                 opts.Source,
		Engine:                 ws.Engine,
		EngineVersion:          opts.EngineVersion,
		Variables:              opts.Variables,
		CreatedBy:              opts.CreatedBy,
		CostEstimationEnabled:  opts.costEstimationEnabled,
	}

	run.Plan = newPhase(run.ID, PlanPhase)
	run.Apply = newPhase(run.ID, ApplyPhase)
	run.updateStatus(runstatus.Pending, opts.now)

	if run.Source == "" {
		run.Source = source.API
	}

	if opts.TerraformVersion != nil {
		run.EngineVersion = *opts.TerraformVersion
	}
	if opts.AllowEmptyApply != nil {
		run.AllowEmptyApply = *opts.AllowEmptyApply
	}
	if opts.IsDestroy != nil {
		run.IsDestroy = *opts.IsDestroy
	}
	if opts.Message != nil {
		run.Message = *opts.Message
	}
	if opts.Refresh != nil {
		run.Refresh = *opts.Refresh
	}
	if opts.AutoApply != nil {
		run.AutoApply = *opts.AutoApply
	}
	if opts.PlanOnly != nil {
		run.PlanOnly = *opts.PlanOnly
	}
	return &run, nil
}

func (run *Run) Queued() bool {
	return runstatus.Queued(run.Status)
}

func (run *Run) HasChanges() bool {
	return run.Plan.HasChanges()
}

// HasStarted is used by the running_time.tmpl partial template to determine
// whether to show the "elapsed time" for a run.
func (run *Run) HasStarted() bool { return true }

// ElapsedTime returns the total time the run has taken thus far. If the run has
// completed, then it is the time taken from entering the pending state
// (creation) through to completion. Otherwise it is the time since entering the
// pending state.
func (run *Run) ElapsedTime(now time.Time) time.Duration {
	pending := run.StatusTimestamps[0]
	if run.Done() {
		completed := run.StatusTimestamps[len(run.StatusTimestamps)-1]
		return completed.Timestamp.Sub(pending.Timestamp)
	}
	return now.Sub(pending.Timestamp)
}

// PeriodReport provides a report of the duration in which a run has been in
// each status thus far. Completed statuses such as completed, errored, etc, are
// ignored because they are an instant not a period of time.
func (run *Run) PeriodReport(now time.Time) (report PeriodReport) {
	// record total time run has taken thus far - it is important that the same
	// 'now' is used both for total time and for the period calculations below
	// so that they add up to the same amounts.
	report.TotalTime = run.ElapsedTime(now)
	if run.Done() {
		// skip last status, which is the completed status
		report.Periods = make([]StatusPeriod, len(run.StatusTimestamps)-1)
	} else {
		report.Periods = make([]StatusPeriod, len(run.StatusTimestamps))
	}
	for i := 0; i < len(run.StatusTimestamps); i++ {
		var (
			duration time.Duration
			current  = run.StatusTimestamps[i]
			isLatest = run.StatusTimestamps[i].Status == run.Status
		)
		if isLatest {
			if run.Done() {
				return
			}
			duration = now.Sub(current.Timestamp)
		} else {
			next := run.StatusTimestamps[i+1]
			duration = next.Timestamp.Sub(current.Timestamp)
		}
		report.Periods[i] = StatusPeriod{
			Status: current.Status,
			Period: duration,
		}
	}
	return report
}

// Phase returns the current phase.
func (run *Run) Phase() PhaseType {
	switch run.Status {
	case runstatus.Pending:
		return PendingPhase
	case runstatus.PlanQueued, runstatus.Planning, runstatus.Planned:
		return PlanPhase
	case runstatus.ApplyQueued, runstatus.Applying, runstatus.Applied:
		return ApplyPhase
	default:
		return UnknownPhase
	}
}

// Discard updates the state of a run to reflect it having been discarded.
func (run *Run) Discard() error {
	if !run.Discardable() {
		return ErrRunDiscardNotAllowed
	}
	run.updateStatus(runstatus.Discarded, nil)

	if run.Status == runstatus.Pending {
		run.Plan.UpdateStatus(PhaseUnreachable)
	}
	run.Apply.UpdateStatus(PhaseUnreachable)

	return nil
}

func (run *Run) InProgress() bool {
	switch run.Status {
	case runstatus.Planning, runstatus.Applying:
		return true
	default:
		return false
	}
}

func (run *Run) String() string {
	return run.ID.String()
}

// Cancel run. Depending upon whether the run is currently in-progress, the run
// is either immediately canceled and its status updated to reflect that, or
// CancelSignaledAt is set to the current time to indicate that a cancelation
// signal should be sent to the process executing the run.
//
// The isUser arg should be set to true if a user is directly instigating the
// cancelation; otherwise it should be set to false, i.e. the job service has
// canceled a job and is now canceling the corresponding run.
//
// The force arg when set to true forceably cancels the run. This is only
// allowed when an attempt has already been made to cancel the run
// non-forceably. The force arg is only respected when isUser is true.
func (run *Run) Cancel(isUser, force bool) error {
	if force {
		if isUser {
			if !run.ForceCancelable() {
				return ErrRunForceCancelNotAllowed
			}
		} else {
			// only a user can forceably cancel a run.
			return ErrRunForceCancelNotAllowed
		}
	}
	var signal bool
	switch run.Status {
	case runstatus.Pending:
		run.Plan.UpdateStatus(PhaseUnreachable)
		run.Apply.UpdateStatus(PhaseUnreachable)
	case runstatus.PlanQueued:
		run.Plan.UpdateStatus(PhaseCanceled)
		run.Apply.UpdateStatus(PhaseUnreachable)
	case runstatus.ApplyQueued:
		run.Apply.UpdateStatus(PhaseCanceled)
	case runstatus.Planning:
		if isUser && !force {
			signal = true
		} else {
			run.Plan.UpdateStatus(PhaseCanceled)
			run.Apply.UpdateStatus(PhaseUnreachable)
		}
	case runstatus.Planned:
		run.Apply.UpdateStatus(PhaseUnreachable)
	case runstatus.Applying:
		if isUser && !force {
			signal = true
		} else {
			run.Apply.UpdateStatus(PhaseCanceled)
		}
	}
	if signal {
		if run.CancelSignaledAt != nil {
			// cannot send cancel signal more than once.
			return ErrRunCancelNotAllowed
		}
		// set timestamp to indicate signal is to be sent, but do not set
		// status to RunCanceled yet.
		now := internal.CurrentTimestamp(nil)
		run.CancelSignaledAt = &now
		return nil
	}
	if force {
		run.updateStatus(runstatus.ForceCanceled, nil)
	} else {
		run.updateStatus(runstatus.Canceled, nil)
	}
	return nil
}

// Cancelable determines whether run can be cancelled.
func (run *Run) Cancelable() bool {
	if run.CancelSignaledAt != nil {
		return false
	}
	switch run.Status {
	case runstatus.Pending, runstatus.PlanQueued, runstatus.Planning, runstatus.ApplyQueued, runstatus.Applying:
		return true
	default:
		return false
	}
}

// ForceCancelable determines whether run can be forceably cancelled.
func (run *Run) ForceCancelable() bool {
	availableAt := run.ForceCancelAvailableAt()
	if availableAt == nil || time.Now().Before(*availableAt) {
		return false
	}
	return true
}

// ForceCancelAvailableAt provides the time from which point it is permitted to
// forceably cancel the run. It only possible to do so when an attempt has
// previously been made to cancel the run non-forceably and a cool-off period
// has elapsed.
func (run *Run) ForceCancelAvailableAt() *time.Time {
	if run.Done() || run.CancelSignaledAt == nil {
		// cannot force cancel a run that is already complete or when no attempt
		// has previously been made to cancel run.
		return nil
	}
	return run.cancelCoolOff()
}

const forceCancelCoolOff = time.Second * 10

func (run *Run) cancelCoolOff() *time.Time {
	if run.CancelSignaledAt == nil {
		return nil
	}
	cooledOff := run.CancelSignaledAt.Add(forceCancelCoolOff)
	return &cooledOff
}

// StartedAt returns the time the run was created.
func (run *Run) StartedAt() time.Time {
	return run.CreatedAt
}

// Done determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (run *Run) Done() bool {
	return runstatus.Done(run.Status)
}

// EnqueuePlan enqueues a plan for the run. It also sets the run as the latest
// run for its workspace (speculative runs are ignored).
func (run *Run) EnqueuePlan() error {
	if run.Status != runstatus.Pending {
		return fmt.Errorf("cannot enqueue run with status %s", run.Status)
	}
	run.updateStatus(runstatus.PlanQueued, nil)
	run.Plan.UpdateStatus(PhaseQueued)

	return nil
}

func (run *Run) EnqueueApply() error {
	switch run.Status {
	case runstatus.Planned, runstatus.CostEstimated:
		// applyable statuses
	default:
		return fmt.Errorf("cannot apply run with status %s", run.Status)
	}
	run.updateStatus(runstatus.ApplyQueued, nil)
	run.Apply.UpdateStatus(PhaseQueued)
	return nil
}

func (run *Run) StatusTimestamp(status runstatus.Status) (time.Time, error) {
	for _, rst := range run.StatusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, internal.ErrStatusTimestampNotFound
}

// Start a run phase
func (run *Run) Start() error {
	switch run.Status {
	case runstatus.PlanQueued:
		run.updateStatus(runstatus.Planning, nil)
		run.Plan.UpdateStatus(PhaseRunning)
	case runstatus.ApplyQueued:
		run.updateStatus(runstatus.Applying, nil)
		run.Apply.UpdateStatus(PhaseRunning)
	case runstatus.Planning, runstatus.Applying:
		return ErrPhaseAlreadyStarted
	default:
		return ErrInvalidRunStateTransition
	}
	return nil
}

// Finish updates the run to reflect its plan or apply phase having finished. If
// a plan phase has finished and an apply should be automatically enqueued then
// autoapply will be set to true.
func (run *Run) Finish(phase PhaseType, opts PhaseFinishOptions) (autoapply bool, err error) {
	if run.Status == runstatus.Canceled {
		// run was canceled before the phase finished so nothing more to do.
		return false, nil
	}
	switch phase {
	case PlanPhase:
		if run.Status != runstatus.Planning {
			return false, ErrInvalidRunStateTransition
		}
		if opts.Errored {
			run.updateStatus(runstatus.Errored, nil)
			run.Plan.UpdateStatus(PhaseErrored)
			run.Apply.UpdateStatus(PhaseUnreachable)
			return false, nil
		}
		// Enter RunCostEstimated state if cost estimation is enabled. OTF does
		// not support cost estimation but enter this state only in order to
		// satisfy the go-tfe tests.
		if run.CostEstimationEnabled {
			run.updateStatus(runstatus.CostEstimated, nil)
		} else {
			run.updateStatus(runstatus.Planned, nil)
		}
		run.Plan.UpdateStatus(PhaseFinished)

		if !run.HasChanges() || run.PlanOnly {
			run.updateStatus(runstatus.PlannedAndFinished, nil)
			run.Apply.UpdateStatus(PhaseUnreachable)
			return false, nil
		}
		return run.AutoApply, nil
	case ApplyPhase:
		if run.Status != runstatus.Applying {
			return false, ErrInvalidRunStateTransition
		}
		if opts.Errored {
			run.updateStatus(runstatus.Errored, nil)
			run.Apply.UpdateStatus(PhaseErrored)
		} else {
			run.updateStatus(runstatus.Applied, nil)
			run.Apply.UpdateStatus(PhaseFinished)
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown phase")
	}
}

func (run *Run) updateStatus(status runstatus.Status, now *time.Time) *Run {
	run.Status = status
	run.StatusTimestamps = append(run.StatusTimestamps, StatusTimestamp{
		Status:    status,
		Timestamp: internal.CurrentTimestamp(now),
	})
	return run
}

// Discardable determines whether run can be discarded.
func (run *Run) Discardable() bool {
	switch run.Status {
	case runstatus.Pending, runstatus.Planned, runstatus.CostEstimated:
		return true
	default:
		return false
	}
}

// Confirmable determines whether run can be confirmed.
func (run *Run) Confirmable() bool {
	switch run.Status {
	case runstatus.Planned:
		return true
	default:
		return false
	}
}

// LogValue implements slog.LogValuer.
func (run *Run) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", run.ID.String()),
		slog.Time("created", run.CreatedAt),
	)
}
