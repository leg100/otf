// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// Run is a terraform run.
type Run struct {
	ID                     resource.TfeID       `jsonapi:"primary,runs"`
	Actions                *RunActions          `jsonapi:"attribute" json:"actions"`
	AllowEmptyApply        bool                 `jsonapi:"attribute" json:"allow-empty-apply"`
	AutoApply              bool                 `jsonapi:"attribute" json:"auto-apply"`
	CreatedAt              time.Time            `jsonapi:"attribute" json:"created-at"`
	ForceCancelAvailableAt *time.Time           `jsonapi:"attribute" json:"force-cancel-available-at"`
	ExecutionMode          string               `jsonapi:"attribute" json:"execution-mode"`
	HasChanges             bool                 `jsonapi:"attribute" json:"has-changes"`
	IsDestroy              bool                 `jsonapi:"attribute" json:"is-destroy"`
	Message                string               `jsonapi:"attribute" json:"message"`
	Permissions            *RunPermissions      `jsonapi:"attribute" json:"permissions"`
	PlanOnly               bool                 `jsonapi:"attribute" json:"plan-only"`
	PositionInQueue        int                  `jsonapi:"attribute" json:"position-in-queue"`
	Refresh                bool                 `jsonapi:"attribute" json:"refresh"`
	RefreshOnly            bool                 `jsonapi:"attribute" json:"refresh-only"`
	ReplaceAddrs           []string             `jsonapi:"attribute" json:"replace-addrs,omitempty"`
	Source                 string               `jsonapi:"attribute" json:"source"`
	Status                 string               `jsonapi:"attribute" json:"status"`
	StatusTimestamps       *RunStatusTimestamps `jsonapi:"attribute" json:"status-timestamps"`
	TargetAddrs            []string             `jsonapi:"attribute" json:"target-addrs,omitempty"`
	TerraformVersion       string               `jsonapi:"attribute" json:"terraform-version"`
	Variables              []RunVariable        `jsonapi:"attribute" json:"variables"`

	// Relations
	Apply                *Apply                `jsonapi:"relationship" json:"apply"`
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relationship" json:"configuration-version"`
	CostEstimate         *CostEstimate         `jsonapi:"relationship" json:"cost-estimate"`
	CreatedBy            *User                 `jsonapi:"relationship" json:"created-by"`
	Plan                 *Plan                 `jsonapi:"relationship" json:"plan"`
	Workspace            *Workspace            `jsonapi:"relationship" json:"workspace"`
}

// RunStatusTimestamps holds the timestamps for individual run statuses.
type RunStatusTimestamps struct {
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
type RunList struct {
	*Pagination
	Items []*Run
}

// RunActions represents the run actions.
type RunActions struct {
	IsCancelable      bool `json:"is-cancelable"`
	IsConfirmable     bool `json:"is-confirmable"`
	IsDiscardable     bool `json:"is-discardable"`
	IsForceCancelable bool `json:"is-force-cancelable"`
}

// RunPermissions represents the run permissions.
type RunPermissions struct {
	CanApply        bool `json:"can-apply"`
	CanCancel       bool `json:"can-cancel"`
	CanDiscard      bool `json:"can-discard"`
	CanForceCancel  bool `json:"can-force-cancel"`
	CanForceExecute bool `json:"can-force-execute"`
}

// RunCreateOptions represents the options for creating a new run.
type RunCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,runs"`

	// AllowEmptyApply specifies whether Terraform can apply the run even when the plan contains no changes.
	// Often used to upgrade state after upgrading a workspace to a new terraform version.
	AllowEmptyApply *bool `jsonapi:"attribute" json:"allow-empty-apply,omitempty"`

	// TerraformVersion specifies the Terraform version to use in this run.
	// Only valid for plan-only runs; must be a valid Terraform version available to the organization.
	TerraformVersion *string `jsonapi:"attribute" json:"terraform-version,omitempty"`

	// PlanOnly specifies if this is a speculative, plan-only run that Terraform cannot apply.
	PlanOnly *bool `jsonapi:"attribute" json:"plan-only,omitempty"`

	// Specifies if this plan is a destroy plan, which will destroy all
	// provisioned resources.
	IsDestroy *bool `jsonapi:"attribute" json:"is-destroy,omitempty"`

	// Refresh determines if the run should
	// update the state prior to checking for differences
	Refresh *bool `jsonapi:"attribute" json:"refresh,omitempty"`

	// RefreshOnly determines whether the run should ignore config changes
	// and refresh the state only
	RefreshOnly *bool `jsonapi:"attribute" json:"refresh-only,omitempty"`

	// Specifies the message to be associated with this run.
	Message *string `jsonapi:"attribute" json:"message,omitempty"`

	// Specifies the configuration version to use for this run. If the
	// configuration version object is omitted, the run will be created using the
	// workspace's latest configuration version.
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relationship" json:"configuration-version"`

	// Specifies the workspace where the run will be executed.
	Workspace *Workspace `jsonapi:"relationship" json:"workspace"`

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
	TargetAddrs []string `jsonapi:"attribute" json:"target-addrs,omitempty"`

	// If non-empty, requests that Terraform create a plan that replaces
	// (destroys and then re-creates) the objects specified by the given
	// resource addresses.
	ReplaceAddrs []string `jsonapi:"attribute" json:"replace-addrs,omitempty"`

	// AutoApply determines if the run should be applied automatically without
	// user confirmation. It defaults to the Workspace.AutoApply setting.
	AutoApply *bool `jsonapi:"attribute" json:"auto-apply,omitempty"`

	// Variables allows you to specify terraform input variables for
	// a particular run, prioritized over variables defined on the workspace.
	Variables []*RunVariable `jsonapi:"attribute" json:"variables,omitempty"`
}

// RunListOptions represents the options for listing runs.
type RunListOptions struct {
	ListOptions

	Organization *resource.OrganizationName `schema:"organization_name,omitempty"`

	WorkspaceID *resource.TfeID `schema:"workspace_id,omitempty"`

	// Optional: Searches runs that matches the supplied VCS username.
	User *string `schema:"search[user],omitempty"`

	// Optional: Searches runs that matches the supplied commit sha.
	Commit *string `schema:"search[commit],omitempty"`

	// Optional: Searches runs that matches the supplied VCS username, commit sha, run_id, and run message.
	// The presence of search[commit] or search[user] takes priority over this parameter and will be omitted.
	Search string `schema:"search[basic],omitempty"`

	// Optional: Comma-separated list of acceptable run statuses.
	// Options are listed at https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#run-states,
	// or as constants with the RunStatus string type.
	Status string `schema:"filter[status],omitempty"`

	// Optional: Comma-separated list of acceptable run sources.
	// Options are listed at https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#run-sources,
	// or as constants with the RunSource string type.
	Source string `schema:"filter[source],omitempty"`

	// Optional: Comma-separated list of acceptable run operation types.
	// Options are listed at https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#run-operations,
	// or as constants with the RunOperation string type.
	Operation string `schema:"filter[operation],omitempty"`

	// Optional: A list of relations to include. See available resources:
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#available-related-resources
	Include []string `schema:"include,omitempty"`
}

// PhaseStatusTimestamps holds the timestamps for individual statuses for a
// phase.
type PhaseStatusTimestamps struct {
	CanceledAt    *time.Time `json:"canceled-at,omitempty"`
	ErroredAt     *time.Time `json:"errored-at,omitempty"`
	FinishedAt    *time.Time `json:"finished-at,omitempty"`
	PendingAt     *time.Time `json:"pending-at,omitempty"`
	QueuedAt      *time.Time `json:"queued-at,omitempty"`
	StartedAt     *time.Time `json:"started-at,omitempty"`
	UnreachableAt *time.Time `json:"unreachable-at,omitempty"`
}

// Apply is a terraform apply
type Apply struct {
	ID               resource.TfeID         `jsonapi:"primary,applies"`
	LogReadURL       string                 `jsonapi:"attribute" json:"log-read-url"`
	Status           string                 `jsonapi:"attribute" json:"status"`
	StatusTimestamps *PhaseStatusTimestamps `jsonapi:"attribute" json:"status-timestamps"`

	ResourceReport
}

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID               resource.TfeID         `jsonapi:"primary,plans"`
	HasChanges       bool                   `jsonapi:"attribute" json:"has-changes"`
	LogReadURL       string                 `jsonapi:"attribute" json:"log-read-url"`
	Status           string                 `jsonapi:"attribute" json:"status"`
	StatusTimestamps *PhaseStatusTimestamps `jsonapi:"attribute" json:"status-timestamps"`

	ResourceReport
}

type ResourceReport struct {
	Additions    *int `json:"resource-additions"`
	Changes      *int `json:"resource-changes"`
	Destructions *int `json:"resource-destructions"`
}

// RunVariable represents a variable that can be applied to a run. All
// values must be expressed as an HCL literal in the same syntax you would use
// when writing terraform code. See
// https://developer.hashicorp.com/terraform/language/expressions/types#types
// for more details.
type RunVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunOperation represents an operation type of run.
type RunOperation string

// List all available run operations.
const (
	RunOperationPlanApply   RunOperation = "plan_and_apply"
	RunOperationPlanOnly    RunOperation = "plan_only"
	RunOperationRefreshOnly RunOperation = "refresh_only"
	RunOperationDestroy     RunOperation = "destroy"
	RunOperationEmptyApply  RunOperation = "empty_apply"
	// **Note: This operation type is still in BETA and subject to change.**
	RunOperationSavePlan RunOperation = "save_plan"
)
