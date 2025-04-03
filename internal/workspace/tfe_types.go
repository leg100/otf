package workspace

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/tfeapi/types"
)

var (
	remoteExecutionMode = "remote"
	localExecutionMode  = "local"
)

// TFEWorkspace represents a Terraform Enterprise workspace.
type TFEWorkspace struct {
	ID                         resource.TfeID        `jsonapi:"primary,workspaces"`
	Actions                    *WorkspaceActions     `jsonapi:"attribute" json:"actions"`
	AgentPoolID                *resource.TfeID       `jsonapi:"attribute" json:"agent-pool-id"`
	AllowDestroyPlan           bool                  `jsonapi:"attribute" json:"allow-destroy-plan"`
	AutoApply                  bool                  `jsonapi:"attribute" json:"auto-apply"`
	CanQueueDestroyPlan        bool                  `jsonapi:"attribute" json:"can-queue-destroy-plan"`
	CreatedAt                  time.Time             `jsonapi:"attribute" json:"created-at"`
	Description                string                `jsonapi:"attribute" json:"description"`
	Environment                string                `jsonapi:"attribute" json:"environment"`
	ExecutionMode              string                `jsonapi:"attribute" json:"execution-mode"`
	FileTriggersEnabled        bool                  `jsonapi:"attribute" json:"file-triggers-enabled"`
	GlobalRemoteState          bool                  `jsonapi:"attribute" json:"global-remote-state"`
	Locked                     bool                  `jsonapi:"attribute" json:"locked"`
	MigrationEnvironment       string                `jsonapi:"attribute" json:"migration-environment"`
	Name                       string                `jsonapi:"attribute" json:"name"`
	Operations                 bool                  `jsonapi:"attribute" json:"operations"`
	Permissions                *WorkspacePermissions `jsonapi:"attribute" json:"permissions"`
	QueueAllRuns               bool                  `jsonapi:"attribute" json:"queue-all-runs"`
	SpeculativeEnabled         bool                  `jsonapi:"attribute" json:"speculative-enabled"`
	SourceName                 string                `jsonapi:"attribute" json:"source-name"`
	SourceURL                  string                `jsonapi:"attribute" json:"source-url"`
	StructuredRunOutputEnabled bool                  `jsonapi:"attribute" json:"structured-run-output-enabled"`
	TerraformVersion           string                `jsonapi:"attribute" json:"terraform-version"`
	TriggerPrefixes            []string              `jsonapi:"attribute" json:"trigger-prefixes"`
	TriggerPatterns            []string              `jsonapi:"attribute" json:"trigger-patterns"`
	VCSRepo                    *VCSRepo              `jsonapi:"attribute" json:"vcs-repo"`
	WorkingDirectory           string                `jsonapi:"attribute" json:"working-directory"`
	UpdatedAt                  time.Time             `jsonapi:"attribute" json:"updated-at"`
	ResourceCount              int                   `jsonapi:"attribute" json:"resource-count"`
	ApplyDurationAverage       time.Duration         `jsonapi:"attribute" json:"apply-duration-average"`
	PlanDurationAverage        time.Duration         `jsonapi:"attribute" json:"plan-duration-average"`
	PolicyCheckFailures        int                   `jsonapi:"attribute" json:"policy-check-failures"`
	RunFailures                int                   `jsonapi:"attribute" json:"run-failures"`
	RunsCount                  int                   `jsonapi:"attribute" json:"workspace-kpis-runs-count"`
	TagNames                   []string              `jsonapi:"attribute" json:"tag-names"`

	// Relations
	CurrentRun   *run.TFERun                   `jsonapi:"relationship" json:"current-run"`
	Organization *organization.TFEOrganization `jsonapi:"relationship" json:"organization"`
	Outputs      []*WorkspaceOutput            `jsonapi:"relationship" json:"outputs"`
}

type WorkspaceOutput struct {
	ID        resource.TfeID `jsonapi:"primary,workspace-outputs"`
	Name      string         `jsonapi:"attribute" json:"name"`
	Sensitive bool           `jsonapi:"attribute" json:"sensitive"`
	Type      string         `jsonapi:"attribute" json:"output-type"`
	Value     any            `jsonapi:"attribute" json:"value"`
}

// WorkspaceList represents a list of workspaces.
type WorkspaceList struct {
	*types.Pagination
	Items []*Workspace
}

// VCSRepo contains the configuration of a VCS integration.
type VCSRepo struct {
	Branch            string         `json:"branch"`
	DisplayIdentifier string         `json:"display-identifier"`
	Identifier        string         `json:"identifier"`
	IngressSubmodules bool           `json:"ingress-submodules"`
	OAuthTokenID      resource.TfeID `json:"oauth-token-id"`
	RepositoryHTTPURL string         `json:"repository-http-url"`
	TagsRegex         string         `json:"tags-regex"`
	ServiceProvider   string         `json:"service-provider"`
}

// WorkspaceActions represents the workspace actions.
type WorkspaceActions struct {
	IsDestroyable bool `json:"is-destroyable"`
}

// WorkspacePermissions represents the workspace permissions.
type WorkspacePermissions struct {
	CanDestroy        bool `json:"can-destroy"`
	CanForceUnlock    bool `json:"can-force-unlock"`
	CanLock           bool `json:"can-lock"`
	CanQueueApply     bool `json:"can-queue-apply"`
	CanQueueDestroy   bool `json:"can-queue-destroy"`
	CanQueueRun       bool `json:"can-queue-run"`
	CanReadSettings   bool `json:"can-read-settings"`
	CanUnlock         bool `json:"can-unlock"`
	CanUpdate         bool `json:"can-update"`
	CanUpdateVariable bool `json:"can-update-variable"`
}

// WorkspaceListOptions represents the options for listing workspaces.
type WorkspaceListOptions struct {
	ListOptions

	// Optional: A search string (partial workspace name) used to filter the results.
	Search string `schema:"search[name],omitempty"`

	// Optional: A search string (comma-separated tag names) used to filter the results.
	Tags string `schema:"search[tags],omitempty"`

	// Optional: A search string (comma-separated tag names to exclude) used to filter the results.
	ExcludeTags string `schema:"search[exclude-tags],omitempty"`

	// Optional: A search on substring matching to filter the results.
	WildcardName string `schema:"search[wildcard-name],omitempty"`

	// Optional: A filter string to list all the workspaces linked to a given project id in the organization.
	ProjectID resource.TfeID `schema:"filter[project][id],omitempty"`

	// Optional: A list of relations to include. See available resources https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	// Include []WSIncludeOpt `url:"include,omitempty"`
}

// WorkspaceCreateOptions represents the options for creating a new workspace.
type WorkspaceCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,workspaces"`

	// Required when execution-mode is set to agent. The ID of the agent pool
	// belonging to the workspace's organization. This value must not be
	// specified if execution-mode is set to remote or local or if operations is
	// set to true.
	AgentPoolID *resource.TfeID `jsonapi:"attribute" json:"agent-pool-id,omitempty"`

	// Whether destroy plans can be queued on the workspace.
	AllowDestroyPlan *bool `jsonapi:"attribute" json:"allow-destroy-plan,omitempty"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attribute" json:"auto-apply,omitempty"`

	// A description for the workspace.
	Description *string `jsonapi:"attribute" json:"description,omitempty"`

	// Which execution mode to use. Valid values are remote, local, and agent.
	// When set to local, the workspace will be used for state storage only.
	// This value must not be specified if operations is specified.
	// 'agent' execution mode is not available in Terraform Enterprise.
	ExecutionMode *string `jsonapi:"attribute" json:"execution-mode,omitempty"`

	// Whether to filter runs based on the changed files in a VCS push. If
	// enabled, the working directory and trigger prefixes describe a set of
	// paths which must contain changes for a VCS push to trigger a run. If
	// disabled, any push will trigger a run.
	FileTriggersEnabled *bool `jsonapi:"attribute" json:"file-triggers-enabled,omitempty"`

	GlobalRemoteState *bool `jsonapi:"attribute" json:"global-remote-state,omitempty"`

	// The legacy TFE environment to use as the source of the migration, in the
	// form organization/environment. Omit this unless you are migrating a legacy
	// environment.
	MigrationEnvironment *string `jsonapi:"attribute" json:"migration-environment,omitempty"`

	// The name of the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// organization.
	Name *string `jsonapi:"attribute" json:"name"`

	// DEPRECATED. Whether the workspace will use remote or local execution mode.
	// Use ExecutionMode instead.
	Operations *bool `jsonapi:"attribute" json:"operations,omitempty"`

	// Organization the workspace belongs to. Required.
	Organization *resource.OrganizationName `schema:"organization_name"`

	// Whether to queue all runs. Unless this is set to true, runs triggered by
	// a webhook will not be queued until at least one run is manually queued.
	QueueAllRuns *bool `jsonapi:"attribute" json:"queue-all-runs,omitempty"`

	// Whether this workspace allows speculative plans. Setting this to false
	// prevents Terraform Cloud or the Terraform Enterprise instance from
	// running plans on pull requests, which can improve security if the VCS
	// repository is public or includes untrusted contributors.
	SpeculativeEnabled *bool `jsonapi:"attribute" json:"speculative-enabled,omitempty"`

	// BETA. A friendly name for the application or client creating this
	// workspace. If set, this will be displayed on the workspace as
	// "Created via <SOURCE NAME>".
	SourceName *string `jsonapi:"attribute" json:"source-name,omitempty"`

	// BETA. A URL for the application or client creating this workspace. This
	// can be the URL of a related resource in another app, or a link to
	// documentation or other info about the client.
	SourceURL *string `jsonapi:"attribute" json:"source-url,omitempty"`

	// BETA. Enable the experimental advanced run user interface.
	// This only applies to runs using Terraform version 0.15.2 or newer,
	// and runs executed using older versions will see the classic experience
	// regardless of this setting.
	StructuredRunOutputEnabled *bool `jsonapi:"attribute" json:"structured-run-output-enabled,omitempty"`

	// The version of Terraform to use for this workspace. Upon creating a
	// workspace, the latest version is selected unless otherwise specified.
	TerraformVersion *string `jsonapi:"attribute" json:"terraform-version,omitempty" schema:"terraform_version"`

	// List of repository-root-relative paths which list all locations to be
	// tracked for changes. See FileTriggersEnabled above for more details.
	TriggerPrefixes []string `jsonapi:"attribute" json:"trigger-prefixes,omitempty"`

	// Optional: List of patterns used to match against changed files in order
	// to decide whether to trigger a run or not.
	TriggerPatterns []string `jsonapi:"attribute" json:"trigger-patterns,omitempty"`

	// Settings for the workspace's VCS repository. If omitted, the workspace is
	// created without a VCS repo. If included, you must specify at least the
	// oauth-token-id and identifier keys below.
	VCSRepo *VCSRepoOptions `jsonapi:"attribute" json:"vcs-repo,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching the
	// environment when multiple environments exist within the same repository.
	WorkingDirectory *string `jsonapi:"attribute" json:"working-directory,omitempty"`

	// A list of tags to attach to the workspace. If the tag does not already
	// exist, it is created and added to the workspace.
	Tags []*Tag `jsonapi:"relationship" json:"tags,omitempty"`
}

// WorkspaceUpdateOptions represents the options for updating a workspace.
type WorkspaceUpdateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,workspaces"`

	// Required when execution-mode is set to agent. The ID of the agent pool
	// belonging to the workspace's organization. This value must not be
	// specified if execution-mode is set to remote or local or if operations is
	// set to true.
	AgentPoolID *resource.TfeID `jsonapi:"attribute" json:"agent-pool-id,omitempty"`

	// Whether destroy plans can be queued on the workspace.
	AllowDestroyPlan *bool `jsonapi:"attribute" json:"allow-destroy-plan,omitempty"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attribute" json:"auto-apply,omitempty"`

	// A new name for the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// organization. Warning: Changing a workspace's name changes its URL in the
	// API and UI.
	Name *string `jsonapi:"attribute" json:"name,omitempty"`

	// A description for the workspace.
	Description *string `jsonapi:"attribute" json:"description,omitempty"`

	// Which execution mode to use. Valid values are remote, local, and agent.
	// When set to local, the workspace will be used for state storage only.
	// This value must not be specified if operations is specified.
	// 'agent' execution mode is not available in Terraform Enterprise.
	ExecutionMode *string `jsonapi:"attribute" json:"execution-mode,omitempty" schema:"execution_mode"`

	// Whether to filter runs based on the changed files in a VCS push. If
	// enabled, the working directory and trigger prefixes describe a set of
	// paths which must contain changes for a VCS push to trigger a run. If
	// disabled, any push will trigger a run.
	FileTriggersEnabled *bool `jsonapi:"attribute" json:"file-triggers-enabled,omitempty"`

	GlobalRemoteState *bool `jsonapi:"attribute" json:"global-remote-state,omitempty"`

	// DEPRECATED. Whether the workspace will use remote or local execution mode.
	// Use ExecutionMode instead.
	Operations *bool `jsonapi:"attribute" json:"operations,omitempty"`

	// Whether to queue all runs. Unless this is set to true, runs triggered by
	// a webhook will not be queued until at least one run is manually queued.
	QueueAllRuns *bool `jsonapi:"attribute" json:"queue-all-runs,omitempty"`

	// Whether this workspace allows speculative plans. Setting this to false
	// prevents Terraform Cloud or the Terraform Enterprise instance from
	// running plans on pull requests, which can improve security if the VCS
	// repository is public or includes untrusted contributors.
	SpeculativeEnabled *bool `jsonapi:"attribute" json:"speculative-enabled,omitempty"`

	// BETA. Enable the experimental advanced run user interface.
	// This only applies to runs using Terraform version 0.15.2 or newer,
	// and runs executed using older versions will see the classic experience
	// regardless of this setting.
	StructuredRunOutputEnabled *bool `jsonapi:"attribute" json:"structured-run-output-enabled,omitempty"`

	// The version of Terraform to use for this workspace.
	TerraformVersion *string `jsonapi:"attribute" json:"terraform-version,omitempty" schema:"terraform_version"`

	// List of repository-root-relative paths which list all locations to be
	// tracked for changes. See FileTriggersEnabled above for more details.
	TriggerPrefixes []string `jsonapi:"attribute" json:"trigger-prefixes,omitempty"`

	// Optional: List of patterns used to match against changed files in order
	// to decide whether to trigger a run or not.
	TriggerPatterns []string `jsonapi:"attribute" json:"trigger-patterns,omitempty"`

	// To delete a workspace's existing VCS repo, specify null instead of an
	// object. To modify a workspace's existing VCS repo, include whichever of
	// the keys below you wish to modify. To add a new VCS repo to a workspace
	// that didn't previously have one, include at least the oauth-token-id and
	// identifier keys.
	VCSRepo VCSRepoOptionsJSON `jsonapi:"attribute" json:"vcs-repo,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching
	// the environment when multiple environments exist within the same
	// repository.
	WorkingDirectory *string `jsonapi:"attribute" json:"working-directory,omitempty"`
}

func (opts *WorkspaceUpdateOptions) Validate() error {
	if opts.Operations != nil && opts.ExecutionMode != nil {
		return errors.New("operations is deprecated and cannot be specified when execution mode is used")
	}
	if opts.Operations != nil {
		if *opts.Operations {
			opts.ExecutionMode = &remoteExecutionMode
		} else {
			opts.ExecutionMode = &localExecutionMode
		}
	}
	return nil
}

// VCSRepoOptions is used by workspaces, policy sets, and registry modules
// VCSRepoOptions represents the configuration options of a VCS integration.
type VCSRepoOptions struct {
	Branch            *string         `json:"branch,omitempty"`
	Identifier        *string         `json:"identifier,omitempty"`
	IngressSubmodules *bool           `json:"ingress-submodules,omitempty"`
	OAuthTokenID      *resource.TfeID `json:"oauth-token-id,omitempty"`
	TagsRegex         *string         `json:"tags-regex,omitempty"`
}

// VCSRepoOptionsJSON wraps VCSRepoOptions and implements json.Unmarshaler in order to differentiate
// between VCSRepoOptions having been explicitly to null, and omitted.
//
// NOTE: Credit to https://www.calhoun.io/how-to-determine-if-a-json-key-has-been-set-to-null-or-not-provided/
type VCSRepoOptionsJSON struct {
	VCSRepoOptions

	Valid bool `json:"-"`
	Set   bool `json:"-"`
}

// UnmarshalJSON differentiates between VCSRepoOptions having been explicitly
// set to null by the client, or the client has left it out.
func (o *VCSRepoOptionsJSON) UnmarshalJSON(data []byte) error {
	// If this method was called, the value was set.
	o.Set = true

	if string(data) == "null" {
		// The key was set to null
		o.Valid = false
		return nil
	}

	// The key isn't set to null
	if err := json.Unmarshal(data, &o.VCSRepoOptions); err != nil {
		return err
	}
	o.Valid = true
	return nil
}
