package run

import (
	"time"

	"github.com/leg100/otf"
)

// Run represents a Terraform Enterprise run.
type jsonapiRun struct {
	ID                     string                      `jsonapi:"primary,runs"`
	Actions                *jsonapiRunActions          `jsonapi:"attr,actions"`
	CreatedAt              time.Time                   `jsonapi:"attr,created-at,iso8601"`
	ForceCancelAvailableAt *time.Time                  `jsonapi:"attr,force-cancel-available-at,iso8601"`
	ExecutionMode          string                      `jsonapi:"attr,execution-mode"`
	HasChanges             bool                        `jsonapi:"attr,has-changes"`
	IsDestroy              bool                        `jsonapi:"attr,is-destroy"`
	Message                string                      `jsonapi:"attr,message"`
	Permissions            *jsonapiRunPermissions      `jsonapi:"attr,permissions"`
	PositionInQueue        int                         `jsonapi:"attr,position-in-queue"`
	Refresh                bool                        `jsonapi:"attr,refresh"`
	RefreshOnly            bool                        `jsonapi:"attr,refresh-only"`
	ReplaceAddrs           []string                    `jsonapi:"attr,replace-addrs,omitempty"`
	Source                 string                      `jsonapi:"attr,source"`
	Status                 string                      `jsonapi:"attr,status"`
	StatusTimestamps       *jsonapiRunStatusTimestamps `jsonapi:"attr,status-timestamps"`
	TargetAddrs            []string                    `jsonapi:"attr,target-addrs,omitempty"`

	// Relations
	Apply                *Apply                `jsonapi:"relation,apply"`
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relation,configuration-version"`
	CreatedBy            otf.User              `jsonapi:"relation,created-by"`
	Plan                 *Plan                 `jsonapi:"relation,plan"`
	Workspace            *Workspace            `jsonapi:"relation,workspace"`
}

// RunStatusTimestamps holds the timestamps for individual run statuses.
type jsonapiRunStatusTimestamps struct {
	AppliedAt            *time.Time `json:"applied-at,omitempty"`
	ApplyQueuedAt        *time.Time `json:"apply-queued-at,omitempty"`
	ApplyingAt           *time.Time `json:"applying-at,omitempty"`
	CanceledAt           *time.Time `json:"canceled-at,omitempty"`
	ConfirmedAt          *time.Time `json:"confirmed-at,omitempty"`
	CostEstimatedAt      *time.Time `json:"cost-estimated-at,omitempty"`
	CostEstimatingAt     *time.Time `json:"cost-estimating-at,omitempty"`
	DiscardedAt          *time.Time `json:"discarded-at,omitempty"`
	ErroredAt            *time.Time `json:"errored-at,omitempty"`
	ForceCanceledAt      *time.Time `json:"force-canceled-at,omitempty"`
	PlanQueueableAt      *time.Time `json:"plan-queueable-at,omitempty"`
	PlanQueuedAt         *time.Time `json:"plan-queued-at,omitempty"`
	PlannedAndFinishedAt *time.Time `json:"planned-and-finished-at,omitempty"`
	PlannedAt            *time.Time `json:"planned-at,omitempty"`
	PlanningAt           *time.Time `json:"planning-at,omitempty"`
	PolicyCheckedAt      *time.Time `json:"policy-checked-at,omitempty"`
	PolicySoftFailedAt   *time.Time `json:"policy-soft-failed-at,omitempty"`
}

// RunList represents a list of runs.
type jsonapiRunList struct {
	*otf.Pagination
	Items []*Run
}

// RunActions represents the run actions.
type jsonapiRunActions struct {
	IsCancelable      bool `json:"is-cancelable"`
	IsConfirmable     bool `json:"is-confirmable"`
	IsDiscardable     bool `json:"is-discardable"`
	IsForceCancelable bool `json:"is-force-cancelable"`
}

// RunPermissions represents the run permissions.
type jsonapiRunPermissions struct {
	CanApply        bool `json:"can-apply"`
	CanCancel       bool `json:"can-cancel"`
	CanDiscard      bool `json:"can-discard"`
	CanForceCancel  bool `json:"can-force-cancel"`
	CanForceExecute bool `json:"can-force-execute"`
}

// RunCreateOptions represents the options for creating a new run.
type jsonapiCreateOptions struct {
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

	// AutoApply determines if the run should be applied automatically without
	// user confirmation. It defaults to the Workspace.AutoApply setting.
	AutoApply *bool `jsonapi:"attr,auto-apply,omitempty"`
}

// Plan represents a Terraform Enterprise plan.
type jsonapiPlan struct {
	ID               string                        `jsonapi:"primary,plans"`
	HasChanges       bool                          `jsonapi:"attr,has-changes"`
	LogReadURL       string                        `jsonapi:"attr,log-read-url"`
	Status           string                        `jsonapi:"attr,status"`
	StatusTimestamps *jsonapiPhaseStatusTimestamps `jsonapi:"attr,status-timestamps"`

	jsonapiResourceReport
}

// Apply represents a Terraform Enterprise apply.
type jsonapiApply struct {
	ID               string                        `jsonapi:"primary,applies"`
	LogReadURL       string                        `jsonapi:"attr,log-read-url"`
	Status           string                        `jsonapi:"attr,status"`
	StatusTimestamps *jsonapiPhaseStatusTimestamps `jsonapi:"attr,status-timestamps"`

	otf.ResourceReport
}

type jsonapiResourceReport struct {
	Additions    *int `json:"resource-additions"`
	Changes      *int `json:"resource-changes"`
	Destructions *int `json:"resource-destructions"`
}

// PhaseStatusTimestamps holds the timestamps for individual statuses for a
// phase.
type jsonapiPhaseStatusTimestamps struct {
	CanceledAt    *time.Time `json:"canceled-at,omitempty"`
	ErroredAt     *time.Time `json:"errored-at,omitempty"`
	FinishedAt    *time.Time `json:"finished-at,omitempty"`
	PendingAt     *time.Time `json:"pending-at,omitempty"`
	QueuedAt      *time.Time `json:"queued-at,omitempty"`
	StartedAt     *time.Time `json:"started-at,omitempty"`
	UnreachableAt *time.Time `json:"unreachable-at,omitempty"`
}
