package otf

import (
	"fmt"
	"time"
)

func (r RunStatus) String() string { return string(r) }

const (
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
	// PlanFormatBinary is the binary representation of the plan file
	PlanFormatBinary = "bin"
	// PlanFormatJSON is the JSON representation of the plan file
	PlanFormatJSON = "json"
)

var (
	ActiveRun = []RunStatus{
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

// PlanFormat is the format of the plan file
type PlanFormat string

func (f PlanFormat) CacheKey(id string) string {
	return fmt.Sprintf("%s.%s", id, f)
}

func (f PlanFormat) SQLColumn() string {
	return fmt.Sprintf("plan_%s", f)
}

type Run interface {
	ID() string
	RunID() string
	CreatedAt() time.Time
	String() string
	IsDestroy() bool
	ForceCancelAvailableAt() *time.Time
	Message() string
	Organization() string
	Refresh() bool
	RefreshOnly() bool
	ReplaceAddrs() []string
	TargetAddrs() []string
	AutoApply() bool
	Speculative() bool
	Status() RunStatus
	WorkspaceID() string
	ConfigurationVersionID() string
	Plan() Plan
	Apply() Apply
	HasChanges() bool
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
	Organization *string `schema:"organization_name,omitempty"`
	// Filter by workspace name
	WorkspaceName *string `schema:"workspace_name,omitempty"`
	// Filter by speculative or non-speculative
	Speculative *bool `schema:"-"`
	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include,omitempty"`
}

type (
	Plan  interface{}
	Apply interface{}
)
