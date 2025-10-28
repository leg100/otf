// Package run is responsible for OTF runs, the primary mechanism for executing
// terraform
package run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
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
		SourceIcon             templ.Component         `json:"-"`
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
		PlanOnly  *bool
		Variables []Variable

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
	}

	// WatchOptions filters events returned by the Watch endpoint.
	WatchOptions struct {
		Organization *organization.Name `schema:"organization_name,omitempty"` // filter by organization name
		WorkspaceID  *resource.TfeID    `schema:"workspace_id,omitempty"`      // filter by workspace ID; mutually exclusive with organization filter
	}

	factoryOrganizationClient interface {
		Get(ctx context.Context, name organization.Name) (*organization.Organization, error)
	}

	factoryWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	}

	factoryConfigClient interface {
		Create(ctx context.Context, workspaceID resource.TfeID, opts configversion.CreateOptions) (*configversion.ConfigurationVersion, error)
		Get(ctx context.Context, id resource.TfeID) (*configversion.ConfigurationVersion, error)
		GetLatest(ctx context.Context, workspaceID resource.TfeID) (*configversion.ConfigurationVersion, error)
		UploadConfig(ctx context.Context, id resource.TfeID, config []byte) error
		GetSourceIcon(source source.Source) templ.Component
	}

	factoryVCSClient interface {
		Get(ctx context.Context, providerID resource.TfeID) (*vcs.Provider, error)
	}

	factoryReleasesClient interface {
		GetLatest(ctx context.Context, engine *engine.Engine) (string, time.Time, error)
	}
)

// NewRun constructs a new run using the provided options.
func NewRun(
	ctx context.Context,
	workspaceID resource.TfeID,
	organizations factoryOrganizationClient,
	workspaces factoryWorkspaceClient,
	configs factoryConfigClient,
	releases factoryReleasesClient,
	vcs factoryVCSClient,
	opts CreateOptions,
) (*Run, error) {
	ws, err := workspaces.Get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	org, err := organizations.Get(ctx, ws.Organization)
	if err != nil {
		return nil, err
	}

	// retrieve or create config: if a config version ID is specified then
	// retrieve that; otherwise if the workspace is connected then the latest
	// config is retrieved from the connected vcs repo, and if the workspace is
	// not connected then the latest existing config is used.
	var cv *configversion.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = configs.Get(ctx, *opts.ConfigurationVersionID)
	} else if ws.Connection == nil {
		cv, err = configs.GetLatest(ctx, workspaceID)
	} else {
		cv, err = createConfigVersionFromVCS(
			ctx,
			ws,
			configs,
			vcs,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("fetching configuration: %w", err)
	}

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
		CostEstimationEnabled:  org.CostEstimationEnabled,
		Source:                 opts.Source,
		Engine:                 ws.Engine,
		Variables:              opts.Variables,
	}
	// If workspace tracks the latest version then fetch it from db.
	if ws.EngineVersion.Latest {
		run.EngineVersion, _, err = releases.GetLatest(ctx, ws.Engine)
		if err != nil {
			return nil, err
		}
	} else {
		run.EngineVersion = ws.EngineVersion.String()
	}

	run.Plan = newPhase(run.ID, PlanPhase)
	run.Apply = newPhase(run.ID, ApplyPhase)
	run.updateStatus(runstatus.Pending, opts.now)

	if run.Source == "" {
		run.Source = source.API
	}
	run.SourceIcon = configs.GetSourceIcon(run.Source)

	if opts.TerraformVersion != nil {
		run.EngineVersion = *opts.TerraformVersion
	}
	if opts.AllowEmptyApply != nil {
		run.AllowEmptyApply = *opts.AllowEmptyApply
	}
	if creator, _ := user.UserFromContext(ctx); creator != nil {
		run.CreatedBy = &creator.Username
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

// createConfigVersionFromVCS creates a config version from the vcs repo
// connected to the workspace using the contents of the vcs repo.
func createConfigVersionFromVCS(
	ctx context.Context,
	ws *workspace.Workspace,
	configs factoryConfigClient,
	vcsClient factoryVCSClient,
) (*configversion.ConfigurationVersion, error) {
	client, err := vcsClient.Get(ctx, ws.Connection.VCSProviderID)
	if err != nil {
		return nil, err
	}
	defaultBranch, err := client.GetDefaultBranch(ctx, ws.Connection.Repo.String())
	if err != nil {
		return nil, fmt.Errorf("retrieving repository info: %w", err)
	}
	branch := ws.Connection.Branch
	if branch == "" {
		branch = defaultBranch
	}
	tarball, ref, err := client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: ws.Connection.Repo,
		Ref:  &branch,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving repository tarball: %w", err)
	}
	commit, err := client.GetCommit(ctx, ws.Connection.Repo, ref)
	if err != nil {
		return nil, fmt.Errorf("retrieving commit information: %s: %w", ref, err)
	}
	cv, err := configs.Create(ctx, ws.ID, configversion.CreateOptions{
		IngressAttributes: &configversion.IngressAttributes{
			Branch:          branch,
			CommitSHA:       commit.SHA,
			CommitURL:       commit.URL,
			Repo:            ws.Connection.Repo,
			IsPullRequest:   false,
			OnDefaultBranch: branch == defaultBranch,
			SenderUsername:  commit.Author.Username,
			SenderAvatarURL: commit.Author.AvatarURL,
			SenderHTMLURL:   commit.Author.ProfileURL,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := configs.UploadConfig(ctx, cv.ID, tarball); err != nil {
		return nil, err
	}
	return cv, err
}

func (r *Run) Queued() bool {
	return runstatus.Queued(r.Status)
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
func (r *Run) PeriodReport(now time.Time) (report PeriodReport) {
	// record total time run has taken thus far - it is important that the same
	// 'now' is used both for total time and for the period calculations below
	// so that they add up to the same amounts.
	report.TotalTime = r.ElapsedTime(now)
	if r.Done() {
		// skip last status, which is the completed status
		report.Periods = make([]StatusPeriod, len(r.StatusTimestamps)-1)
	} else {
		report.Periods = make([]StatusPeriod, len(r.StatusTimestamps))
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
		report.Periods[i] = StatusPeriod{
			Status: current.Status,
			Period: duration,
		}
	}
	return report
}

// Phase returns the current phase.
func (r *Run) Phase() PhaseType {
	switch r.Status {
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
func (r *Run) Discard() error {
	if !r.Discardable() {
		return ErrRunDiscardNotAllowed
	}
	r.updateStatus(runstatus.Discarded, nil)

	if r.Status == runstatus.Pending {
		r.Plan.UpdateStatus(PhaseUnreachable)
	}
	r.Apply.UpdateStatus(PhaseUnreachable)

	return nil
}

func (r *Run) InProgress() bool {
	switch r.Status {
	case runstatus.Planning, runstatus.Applying:
		return true
	default:
		return false
	}
}

func (r *Run) String() string {
	return r.ID.String()
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
func (r *Run) Cancel(isUser, force bool) error {
	if force {
		if isUser {
			if !r.ForceCancelable() {
				return ErrRunForceCancelNotAllowed
			}
		} else {
			// only a user can forceably cancel a run.
			return ErrRunForceCancelNotAllowed
		}
	}
	var signal bool
	switch r.Status {
	case runstatus.Pending:
		r.Plan.UpdateStatus(PhaseUnreachable)
		r.Apply.UpdateStatus(PhaseUnreachable)
	case runstatus.PlanQueued:
		r.Plan.UpdateStatus(PhaseCanceled)
		r.Apply.UpdateStatus(PhaseUnreachable)
	case runstatus.ApplyQueued:
		r.Apply.UpdateStatus(PhaseCanceled)
	case runstatus.Planning:
		if isUser && !force {
			signal = true
		} else {
			r.Plan.UpdateStatus(PhaseCanceled)
			r.Apply.UpdateStatus(PhaseUnreachable)
		}
	case runstatus.Planned:
		r.Apply.UpdateStatus(PhaseUnreachable)
	case runstatus.Applying:
		if isUser && !force {
			signal = true
		} else {
			r.Apply.UpdateStatus(PhaseCanceled)
		}
	}
	if signal {
		if r.CancelSignaledAt != nil {
			// cannot send cancel signal more than once.
			return ErrRunCancelNotAllowed
		}
		// set timestamp to indicate signal is to be sent, but do not set
		// status to RunCanceled yet.
		now := internal.CurrentTimestamp(nil)
		r.CancelSignaledAt = &now
		return nil
	}
	if force {
		r.updateStatus(runstatus.ForceCanceled, nil)
	} else {
		r.updateStatus(runstatus.Canceled, nil)
	}
	return nil
}

// Cancelable determines whether run can be cancelled.
func (r *Run) Cancelable() bool {
	if r.CancelSignaledAt != nil {
		return false
	}
	switch r.Status {
	case runstatus.Pending, runstatus.PlanQueued, runstatus.Planning, runstatus.ApplyQueued, runstatus.Applying:
		return true
	default:
		return false
	}
}

// ForceCancelable determines whether run can be forceably cancelled.
func (r *Run) ForceCancelable() bool {
	availableAt := r.ForceCancelAvailableAt()
	if availableAt == nil || time.Now().Before(*availableAt) {
		return false
	}
	return true
}

// ForceCancelAvailableAt provides the time from which point it is permitted to
// forceably cancel the run. It only possible to do so when an attempt has
// previously been made to cancel the run non-forceably and a cool-off period
// has elapsed.
func (r *Run) ForceCancelAvailableAt() *time.Time {
	if r.Done() || r.CancelSignaledAt == nil {
		// cannot force cancel a run that is already complete or when no attempt
		// has previously been made to cancel run.
		return nil
	}
	return r.cancelCoolOff()
}

const forceCancelCoolOff = time.Second * 10

func (r *Run) cancelCoolOff() *time.Time {
	if r.CancelSignaledAt == nil {
		return nil
	}
	cooledOff := r.CancelSignaledAt.Add(forceCancelCoolOff)
	return &cooledOff
}

// StartedAt returns the time the run was created.
func (r *Run) StartedAt() time.Time {
	return r.CreatedAt
}

// Done determines whether run has reached an end state, e.g. applied,
// discarded, etc.
func (r *Run) Done() bool {
	return runstatus.Done(r.Status)
}

// EnqueuePlan enqueues a plan for the run. It also sets the run as the latest
// run for its workspace (speculative runs are ignored).
func (r *Run) EnqueuePlan() error {
	if r.Status != runstatus.Pending {
		return fmt.Errorf("cannot enqueue run with status %s", r.Status)
	}
	r.updateStatus(runstatus.PlanQueued, nil)
	r.Plan.UpdateStatus(PhaseQueued)

	return nil
}

func (r *Run) EnqueueApply() error {
	switch r.Status {
	case runstatus.Planned, runstatus.CostEstimated:
		// applyable statuses
	default:
		return fmt.Errorf("cannot apply run with status %s", r.Status)
	}
	r.updateStatus(runstatus.ApplyQueued, nil)
	r.Apply.UpdateStatus(PhaseQueued)
	return nil
}

func (r *Run) StatusTimestamp(status runstatus.Status) (time.Time, error) {
	for _, rst := range r.StatusTimestamps {
		if rst.Status == status {
			return rst.Timestamp, nil
		}
	}
	return time.Time{}, internal.ErrStatusTimestampNotFound
}

// Start a run phase
func (r *Run) Start() error {
	switch r.Status {
	case runstatus.PlanQueued:
		r.updateStatus(runstatus.Planning, nil)
		r.Plan.UpdateStatus(PhaseRunning)
	case runstatus.ApplyQueued:
		r.updateStatus(runstatus.Applying, nil)
		r.Apply.UpdateStatus(PhaseRunning)
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
func (r *Run) Finish(phase PhaseType, opts PhaseFinishOptions) (autoapply bool, err error) {
	if r.Status == runstatus.Canceled {
		// run was canceled before the phase finished so nothing more to do.
		return false, nil
	}
	switch phase {
	case PlanPhase:
		if r.Status != runstatus.Planning {
			return false, ErrInvalidRunStateTransition
		}
		if opts.Errored {
			r.updateStatus(runstatus.Errored, nil)
			r.Plan.UpdateStatus(PhaseErrored)
			r.Apply.UpdateStatus(PhaseUnreachable)
			return false, nil
		}
		// Enter RunCostEstimated state if cost estimation is enabled. OTF does
		// not support cost estimation but enter this state only in order to
		// satisfy the go-tfe tests.
		if r.CostEstimationEnabled {
			r.updateStatus(runstatus.CostEstimated, nil)
		} else {
			r.updateStatus(runstatus.Planned, nil)
		}
		r.Plan.UpdateStatus(PhaseFinished)

		if !r.HasChanges() || r.PlanOnly {
			r.updateStatus(runstatus.PlannedAndFinished, nil)
			r.Apply.UpdateStatus(PhaseUnreachable)
			return false, nil
		}
		return r.AutoApply, nil
	case ApplyPhase:
		if r.Status != runstatus.Applying {
			return false, ErrInvalidRunStateTransition
		}
		if opts.Errored {
			r.updateStatus(runstatus.Errored, nil)
			r.Apply.UpdateStatus(PhaseErrored)
		} else {
			r.updateStatus(runstatus.Applied, nil)
			r.Apply.UpdateStatus(PhaseFinished)
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown phase")
	}
}

func (r *Run) updateStatus(status runstatus.Status, now *time.Time) *Run {
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
	case runstatus.Pending, runstatus.Planned, runstatus.CostEstimated:
		return true
	default:
		return false
	}
}

// Confirmable determines whether run can be confirmed.
func (r *Run) Confirmable() bool {
	switch r.Status {
	case runstatus.Planned:
		return true
	default:
		return false
	}
}
