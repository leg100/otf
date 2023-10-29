// Package run is responsible for OTF runs, the primary mechanism for executing
// terraform
package run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
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
		ID                     string                  `jsonapi:"primary,runs"`
		CreatedAt              time.Time               `jsonapi:"attribute" json:"created_at"`
		IsDestroy              bool                    `jsonapi:"attribute" json:"is_destroy"`
		ForceCancelAvailableAt *time.Time              `jsonapi:"attribute" json:"force_cancel_available_at"`
		Message                string                  `jsonapi:"attribute" json:"message"`
		Organization           string                  `jsonapi:"attribute" json:"organization"`
		Refresh                bool                    `jsonapi:"attribute" json:"refresh"`
		RefreshOnly            bool                    `jsonapi:"attribute" json:"refresh_only"`
		ReplaceAddrs           []string                `jsonapi:"attribute" json:"replace_addrs"`
		PositionInQueue        int                     `jsonapi:"attribute" json:"position_in_queue"`
		TargetAddrs            []string                `jsonapi:"attribute" json:"target_addrs"`
		TerraformVersion       string                  `jsonapi:"attribute" json:"terraform_version"`
		AllowEmptyApply        bool                    `jsonapi:"attribute" json:"allow_empty_apply"`
		AutoApply              bool                    `jsonapi:"attribute" json:"auto_apply"`
		PlanOnly               bool                    `jsonapi:"attribute" json:"plan_only"`
		Source                 Source                  `jsonapi:"attribute" json:"source"`
		Status                 internal.RunStatus      `jsonapi:"attribute" json:"status"`
		WorkspaceID            string                  `jsonapi:"attribute" json:"workspace_id"`
		ConfigurationVersionID string                  `jsonapi:"attribute" json:"configuration_version_id"`
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
		CreatedBy *string

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
		Status    internal.RunStatus `json:"status"`
		Timestamp time.Time          `json:"timestamp"`
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
		ConfigurationVersionID *string
		TargetAddrs            []string
		ReplaceAddrs           []string
		AutoApply              *bool
		Source                 Source
		TerraformVersion       *string
		AllowEmptyApply        *bool
		// PlanOnly specifies if this is a speculative, plan-only run that
		// Terraform cannot apply. Takes precedence over whether the
		// configuration version is marked as speculative or not.
		PlanOnly  *bool
		Variables []Variable

		// testing purposes
		now *time.Time
	}

	// ListOptions are options for paginating and filtering a list of runs
	ListOptions struct {
		resource.PageOptions
		// Filter by workspace ID
		WorkspaceID *string `schema:"workspace_id,omitempty"`
		// Filter by organization name
		Organization *string `schema:"organization_name,omitempty"`
		// Filter by workspace name
		WorkspaceName *string `schema:"workspace_name,omitempty"`
		// Filter by run statuses (with an implicit OR condition)
		Statuses []internal.RunStatus `schema:"statuses,omitempty"`
		// Filter by plan-only runs
		PlanOnly *bool `schema:"-"`
		// Filter by sources
		Sources []Source
		// Filter by commit SHA that triggered a run
		CommitSHA *string
		// Filter by VCS user's username that triggered a run
		VCSUsername *string
	}

	// WatchOptions filters events returned by the Watch endpoint.
	WatchOptions struct {
		Organization *string `schema:"organization_name,omitempty"` // filter by organization name
		WorkspaceID  *string `schema:"workspace_id,omitempty"`      // filter by workspace ID; mutually exclusive with organization filter
	}
)

// newRun creates a new run with defaults.
func newRun(ctx context.Context, org *organization.Organization, cv *configversion.ConfigurationVersion, ws *workspace.Workspace, opts CreateOptions) *Run {
	run := Run{
		ID:                     internal.NewID("run"),
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
		CostEstimationEnabled:  org.CostEstimationEnabled,
		Source:                 opts.Source,
		TerraformVersion:       ws.TerraformVersion,
		Variables:              opts.Variables,
	}
	run.Plan = newPhase(run.ID, internal.PlanPhase)
	run.Apply = newPhase(run.ID, internal.ApplyPhase)
	run.updateStatus(internal.RunPending, opts.now)

	if run.Source == "" {
		run.Source = SourceAPI
	}
	if opts.TerraformVersion != nil {
		run.TerraformVersion = *opts.TerraformVersion
	}
	if opts.AllowEmptyApply != nil {
		run.AllowEmptyApply = *opts.AllowEmptyApply
	}
	if user, _ := auth.UserFromContext(ctx); user != nil {
		run.CreatedBy = &user.Username
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
	return &run
}

func (r *Run) String() string { return r.ID }

func (r *Run) Queued() bool {
	return r.Status == internal.RunPlanQueued || r.Status == internal.RunApplyQueued
}

func (r *Run) HasChanges() bool {
	return r.Plan.HasChanges()
}

// HasStarted is used by the running_time.tmpl partial template to determine
// whether to show the "elapsed time" for a run.
func (r *Run) HasStarted() bool { return true }

// ElapsedTime returns the total time the run has taken thus far. If the run has
// completed, then it is the time taken from entering the pending state
// (creation) through to completion. Otherwise it is the time since entering the
// pending state.
func (r *Run) ElapsedTime(now time.Time) time.Duration {
	pending := r.StatusTimestamps[0]
	if r.Done() {
		completed := r.StatusTimestamps[len(r.StatusTimestamps)-1]
		return completed.Timestamp.Sub(pending.Timestamp)
	}
	return now.Sub(pending.Timestamp)
}

// PeriodReport provides a report of the duration in which a run has been in
// each status thus far. Completed statuses such as completed, errored, etc, are
// ignored because they are an instant not a period of time.
func (r *Run) PeriodReport(now time.Time) (report internal.PeriodReport) {
	// record total time run has taken thus far - it is important that the same
	// 'now' is used both for total time and for the period calculations below
	// so that they add up to the same amounts.
	report.TotalTime = r.ElapsedTime(now)
	if r.Done() {
		// skip last status, which is the completed status
		report.Periods = make([]internal.StatusPeriod, len(r.StatusTimestamps)-1)
	} else {
		report.Periods = make([]internal.StatusPeriod, len(r.StatusTimestamps))
	}
	for i := 0; i < len(r.StatusTimestamps); i++ {
		var (
			duration time.Duration
			current  = r.StatusTimestamps[i]
			isLatest = r.StatusTimestamps[i].Status == r.Status
		)
		if isLatest {
			if r.Done() {
				return
			}
			duration = now.Sub(current.Timestamp)
		} else {
			next := r.StatusTimestamps[i+1]
			duration = next.Timestamp.Sub(current.Timestamp)
		}
		report.Periods[i] = internal.StatusPeriod{
			Status: current.Status,
			Period: duration,
		}
	}
	return report
}

// Phase returns the current phase.
func (r *Run) Phase() internal.PhaseType {
	switch r.Status {
	case internal.RunPending:
		return internal.PendingPhase
	case internal.RunPlanQueued, internal.RunPlanning, internal.RunPlanned:
		return internal.PlanPhase
	case internal.RunApplyQueued, internal.RunApplying, internal.RunApplied:
		return internal.ApplyPhase
	default:
		return internal.UnknownPhase
	}
}

// Discard updates the state of a run to reflect it having been discarded.
func (r *Run) Discard() error {
	if !r.Discardable() {
		return internal.ErrRunDiscardNotAllowed
	}
	r.updateStatus(internal.RunDiscarded, nil)

	if r.Status == internal.RunPending {
		r.Plan.UpdateStatus(PhaseUnreachable)
	}
	r.Apply.UpdateStatus(PhaseUnreachable)

	return nil
}

// Cancel run. Returns a boolean indicating whether a cancel request should be
// enqueued (for an agent to kill an in progress process)
func (r *Run) Cancel() error {
	if !r.Cancelable() {
		return internal.ErrRunCancelNotAllowed
	}
	// permit run to be force canceled after a cool off period of 10 seconds has
	// elapsed.
	tenSecondsFromNow := internal.CurrentTimestamp(nil).Add(10 * time.Second)
	r.ForceCancelAvailableAt = &tenSecondsFromNow

	switch r.Status {
	case internal.RunPending:
		r.Plan.UpdateStatus(PhaseUnreachable)
		r.Apply.UpdateStatus(PhaseUnreachable)
	case internal.RunPlanQueued, internal.RunPlanning:
		r.Plan.UpdateStatus(PhaseCanceled)
		r.Apply.UpdateStatus(PhaseUnreachable)
	case internal.RunApplyQueued, internal.RunApplying:
		r.Apply.UpdateStatus(PhaseCanceled)
	}

	r.updateStatus(internal.RunCanceled, nil)

	return nil
}

// ForceCancel force cancels a run. A cool-off period of 10 seconds must have
// elapsed following a cancelation request before a run can be force canceled.
func (r *Run) ForceCancel() error {
	if r.ForceCancelAvailableAt != nil && time.Now().After(*r.ForceCancelAvailableAt) {
		r.updateStatus(internal.RunForceCanceled, nil)
		return nil
	}
	return internal.ErrRunForceCancelNotAllowed
}

// StartedAt returns the time the run was created.
func (r *Run) StartedAt() time.Time {
	return r.CreatedAt
}

// Done determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (r *Run) Done() bool {
	switch r.Status {
	case internal.RunApplied, internal.RunPlannedAndFinished, internal.RunDiscarded, internal.RunCanceled, internal.RunErrored:
		return true
	default:
		return false
	}
}

// EnqueuePlan enqueues a plan for the run. It also sets the run as the latest
// run for its workspace (speculative runs are ignored).
func (r *Run) EnqueuePlan() error {
	if r.Status != internal.RunPending {
		return fmt.Errorf("cannot enqueue run with status %s", r.Status)
	}
	r.updateStatus(internal.RunPlanQueued, nil)
	r.Plan.UpdateStatus(PhaseQueued)

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

func (r *Run) CanAccessWorkspace(action rbac.Action, policy *internal.WorkspacePolicy) bool {
	// run can access anything within its workspace
	return r.WorkspaceID == policy.WorkspaceID
}

func (r *Run) EnqueueApply() error {
	switch r.Status {
	case internal.RunPlanned, internal.RunCostEstimated:
		// applyable statuses
	default:
		return fmt.Errorf("cannot apply run with status %s", r.Status)
	}
	r.updateStatus(internal.RunApplyQueued, nil)
	r.Apply.UpdateStatus(PhaseQueued)
	return nil
}

func (r *Run) StatusTimestamp(status internal.RunStatus) (time.Time, error) {
	for _, rst := range r.StatusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, internal.ErrStatusTimestampNotFound
}

// Start a run phase
func (r *Run) Start(phase internal.PhaseType) error {
	switch r.Status {
	case internal.RunPlanQueued:
		r.updateStatus(internal.RunPlanning, nil)
		r.Plan.UpdateStatus(PhaseRunning)
	case internal.RunApplyQueued:
		r.updateStatus(internal.RunApplying, nil)
		r.Apply.UpdateStatus(PhaseRunning)
	case internal.RunPlanning, internal.RunApplying:
		return internal.ErrPhaseAlreadyStarted
	default:
		return ErrInvalidRunStateTransition
	}
	return nil
}

// Finish updates the run to reflect its plan or apply phase having finished.
func (r *Run) Finish(phase internal.PhaseType, opts PhaseFinishOptions) error {
	if r.Status == internal.RunCanceled {
		// run was canceled before the phase finished so nothing more to do.
		return nil
	}
	switch phase {
	case internal.PlanPhase:
		if r.Status != internal.RunPlanning {
			return ErrInvalidRunStateTransition
		}
		if opts.Errored {
			r.updateStatus(internal.RunErrored, nil)
			r.Plan.UpdateStatus(PhaseErrored)
			r.Apply.UpdateStatus(PhaseUnreachable)
			return nil
		}
		// Enter RunCostEstimated state if cost estimation is enabled. OTF does
		// not support cost estimation but enter this state only in order to
		// satisfy the go-tfe tests.
		if r.CostEstimationEnabled {
			r.updateStatus(internal.RunCostEstimated, nil)
		} else {
			r.updateStatus(internal.RunPlanned, nil)
		}
		r.Plan.UpdateStatus(PhaseFinished)

		if !r.HasChanges() || r.PlanOnly {
			r.updateStatus(internal.RunPlannedAndFinished, nil)
			r.Apply.UpdateStatus(PhaseUnreachable)
		} else if r.AutoApply {
			return r.EnqueueApply()
		}
		return nil
	case internal.ApplyPhase:
		if r.Status != internal.RunApplying {
			return ErrInvalidRunStateTransition
		}
		if opts.Errored {
			r.updateStatus(internal.RunErrored, nil)
			r.Apply.UpdateStatus(PhaseErrored)
		} else {
			r.updateStatus(internal.RunApplied, nil)
			r.Apply.UpdateStatus(PhaseFinished)
		}
		return nil
	default:
		return fmt.Errorf("unknown phase")
	}
}

func (r *Run) updateStatus(status internal.RunStatus, now *time.Time) *Run {
	r.Status = status
	r.StatusTimestamps = append(r.StatusTimestamps, StatusTimestamp{
		Status:    status,
		Timestamp: internal.CurrentTimestamp(now),
	})
	return r
}

// Discardable determines whether run can be discarded.
func (r *Run) Discardable() bool {
	switch r.Status {
	case internal.RunPending, internal.RunPlanned, internal.RunCostEstimated:
		return true
	default:
		return false
	}
}

// Cancelable determines whether run can be cancelled.
func (r *Run) Cancelable() bool {
	switch r.Status {
	case internal.RunPending, internal.RunPlanQueued, internal.RunPlanning, internal.RunApplyQueued, internal.RunApplying:
		return true
	default:
		return false
	}
}

// Confirmable determines whether run can be confirmed.
func (r *Run) Confirmable() bool {
	switch r.Status {
	case internal.RunPlanned:
		return true
	default:
		return false
	}
}

// Helper methods for templates; helps avoid using strings within templates to refer
// to constants.

func (r *Run) IsGithubSource() bool { return r.Source == SourceGithub }
func (r *Run) IsGitlabSource() bool { return r.Source == SourceGitlab }
func (r *Run) IsUISource() bool     { return r.Source == SourceUI }
func (r *Run) IsAPISource() bool    { return r.Source == SourceAPI }
func (r *Run) IsCLISource() bool    { return r.Source == SourceTerraform }
