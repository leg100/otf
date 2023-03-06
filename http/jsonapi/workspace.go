package jsonapi

import (
	"errors"
	"time"
)

var (
	remoteExecutionMode = "remote"
	localExecutionMode  = "local"
)

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	ID                         string                `jsonapi:"primary,workspaces"`
	Actions                    *WorkspaceActions     `jsonapi:"attr,actions"`
	AgentPoolID                string                `jsonapi:"attr,agent-pool-id"`
	AllowDestroyPlan           bool                  `jsonapi:"attr,allow-destroy-plan"`
	AutoApply                  bool                  `jsonapi:"attr,auto-apply"`
	CanQueueDestroyPlan        bool                  `jsonapi:"attr,can-queue-destroy-plan"`
	CreatedAt                  time.Time             `jsonapi:"attr,created-at,iso8601"`
	Description                string                `jsonapi:"attr,description"`
	Environment                string                `jsonapi:"attr,environment"`
	ExecutionMode              string                `jsonapi:"attr,execution-mode"`
	FileTriggersEnabled        bool                  `jsonapi:"attr,file-triggers-enabled"`
	GlobalRemoteState          bool                  `jsonapi:"attr,global-remote-state"`
	Locked                     bool                  `jsonapi:"attr,locked"`
	MigrationEnvironment       string                `jsonapi:"attr,migration-environment"`
	Name                       string                `jsonapi:"attr,name"`
	Operations                 bool                  `jsonapi:"attr,operations"`
	Permissions                *WorkspacePermissions `jsonapi:"attr,permissions"`
	QueueAllRuns               bool                  `jsonapi:"attr,queue-all-runs"`
	SpeculativeEnabled         bool                  `jsonapi:"attr,speculative-enabled"`
	SourceName                 string                `jsonapi:"attr,source-name"`
	SourceURL                  string                `jsonapi:"attr,source-url"`
	StructuredRunOutputEnabled bool                  `jsonapi:"attr,structured-run-output-enabled"`
	TerraformVersion           string                `jsonapi:"attr,terraform-version"`
	TriggerPrefixes            []string              `jsonapi:"attr,trigger-prefixes"`
	VCSRepo                    *VCSRepo              `jsonapi:"attr,vcs-repo"`
	WorkingDirectory           string                `jsonapi:"attr,working-directory"`
	UpdatedAt                  time.Time             `jsonapi:"attr,updated-at,iso8601"`
	ResourceCount              int                   `jsonapi:"attr,resource-count"`
	ApplyDurationAverage       time.Duration         `jsonapi:"attr,apply-duration-average"`
	PlanDurationAverage        time.Duration         `jsonapi:"attr,plan-duration-average"`
	PolicyCheckFailures        int                   `jsonapi:"attr,policy-check-failures"`
	RunFailures                int                   `jsonapi:"attr,run-failures"`
	RunsCount                  int                   `jsonapi:"attr,workspace-kpis-runs-count"`

	// Relations
	CurrentRun   *Run          `jsonapi:"relation,current-run"`
	Organization *Organization `jsonapi:"relation,organization"`
}

// WorkspaceList represents a list of workspaces.
type WorkspaceList struct {
	*Pagination
	Items []*Workspace
}

// VCSRepo contains the configuration of a VCS integration.
type VCSRepo struct {
	Branch            string `json:"branch"`
	DisplayIdentifier string `json:"display-identifier"`
	Identifier        string `json:"identifier"`
	IngressSubmodules bool   `json:"ingress-submodules"`
	OAuthTokenID      string `json:"oauth-token-id"`
	RepositoryHTTPURL string `json:"repository-http-url"`
	ServiceProvider   string `json:"service-provider"`
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
	AgentPoolID *string `jsonapi:"attr,agent-pool-id,omitempty"`

	// Whether destroy plans can be queued on the workspace.
	AllowDestroyPlan *bool `jsonapi:"attr,allow-destroy-plan,omitempty"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attr,auto-apply,omitempty"`

	// A description for the workspace.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Which execution mode to use. Valid values are remote, local, and agent.
	// When set to local, the workspace will be used for state storage only.
	// This value must not be specified if operations is specified.
	// 'agent' execution mode is not available in Terraform Enterprise.
	ExecutionMode *string `jsonapi:"attr,execution-mode,omitempty"`

	// Whether to filter runs based on the changed files in a VCS push. If
	// enabled, the working directory and trigger prefixes describe a set of
	// paths which must contain changes for a VCS push to trigger a run. If
	// disabled, any push will trigger a run.
	FileTriggersEnabled *bool `jsonapi:"attr,file-triggers-enabled,omitempty"`

	GlobalRemoteState *bool `jsonapi:"attr,global-remote-state,omitempty"`

	// The legacy TFE environment to use as the source of the migration, in the
	// form organization/environment. Omit this unless you are migrating a legacy
	// environment.
	MigrationEnvironment *string `jsonapi:"attr,migration-environment,omitempty"`

	// The name of the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// organization.
	Name *string `jsonapi:"attr,name"`

	// DEPRECATED. Whether the workspace will use remote or local execution mode.
	// Use ExecutionMode instead.
	Operations *bool `jsonapi:"attr,operations,omitempty"`

	// Organization the workspace belongs to. Required.
	Organization *string `schema:"organization_name"`

	// Whether to queue all runs. Unless this is set to true, runs triggered by
	// a webhook will not be queued until at least one run is manually queued.
	QueueAllRuns *bool `jsonapi:"attr,queue-all-runs,omitempty"`

	// Whether this workspace allows speculative plans. Setting this to false
	// prevents Terraform Cloud or the Terraform Enterprise instance from
	// running plans on pull requests, which can improve security if the VCS
	// repository is public or includes untrusted contributors.
	SpeculativeEnabled *bool `jsonapi:"attr,speculative-enabled,omitempty"`

	// BETA. A friendly name for the application or client creating this
	// workspace. If set, this will be displayed on the workspace as
	// "Created via <SOURCE NAME>".
	SourceName *string `jsonapi:"attr,source-name,omitempty"`

	// BETA. A URL for the application or client creating this workspace. This
	// can be the URL of a related resource in another app, or a link to
	// documentation or other info about the client.
	SourceURL *string `jsonapi:"attr,source-url,omitempty"`

	// BETA. Enable the experimental advanced run user interface.
	// This only applies to runs using Terraform version 0.15.2 or newer,
	// and runs executed using older versions will see the classic experience
	// regardless of this setting.
	StructuredRunOutputEnabled *bool `jsonapi:"attr,structured-run-output-enabled,omitempty"`

	// The version of Terraform to use for this workspace. Upon creating a
	// workspace, the latest version is selected unless otherwise specified.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty" schema:"terraform_version"`

	// List of repository-root-relative paths which list all locations to be
	// tracked for changes. See FileTriggersEnabled above for more details.
	TriggerPrefixes []string `jsonapi:"attr,trigger-prefixes,omitempty"`

	// Settings for the workspace's VCS repository. If omitted, the workspace is
	// created without a VCS repo. If included, you must specify at least the
	// oauth-token-id and identifier keys below.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching the
	// environment when multiple environments exist within the same repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`
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
	AgentPoolID *string `jsonapi:"attr,agent-pool-id,omitempty"`

	// Whether destroy plans can be queued on the workspace.
	AllowDestroyPlan *bool `jsonapi:"attr,allow-destroy-plan,omitempty"`

	// Whether to automatically apply changes when a Terraform plan is successful.
	AutoApply *bool `jsonapi:"attr,auto-apply,omitempty"`

	// A new name for the workspace, which can only include letters, numbers, -,
	// and _. This will be used as an identifier and must be unique in the
	// organization. Warning: Changing a workspace's name changes its URL in the
	// API and UI.
	Name *string `jsonapi:"attr,name,omitempty"`

	// A description for the workspace.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Which execution mode to use. Valid values are remote, local, and agent.
	// When set to local, the workspace will be used for state storage only.
	// This value must not be specified if operations is specified.
	// 'agent' execution mode is not available in Terraform Enterprise.
	ExecutionMode *string `jsonapi:"attr,execution-mode,omitempty" schema:"execution_mode"`

	// Whether to filter runs based on the changed files in a VCS push. If
	// enabled, the working directory and trigger prefixes describe a set of
	// paths which must contain changes for a VCS push to trigger a run. If
	// disabled, any push will trigger a run.
	FileTriggersEnabled *bool `jsonapi:"attr,file-triggers-enabled,omitempty"`

	GlobalRemoteState *bool `jsonapi:"attr,global-remote-state,omitempty"`

	// DEPRECATED. Whether the workspace will use remote or local execution mode.
	// Use ExecutionMode instead.
	Operations *bool `jsonapi:"attr,operations,omitempty"`

	// Whether to queue all runs. Unless this is set to true, runs triggered by
	// a webhook will not be queued until at least one run is manually queued.
	QueueAllRuns *bool `jsonapi:"attr,queue-all-runs,omitempty"`

	// Whether this workspace allows speculative plans. Setting this to false
	// prevents Terraform Cloud or the Terraform Enterprise instance from
	// running plans on pull requests, which can improve security if the VCS
	// repository is public or includes untrusted contributors.
	SpeculativeEnabled *bool `jsonapi:"attr,speculative-enabled,omitempty"`

	// BETA. Enable the experimental advanced run user interface.
	// This only applies to runs using Terraform version 0.15.2 or newer,
	// and runs executed using older versions will see the classic experience
	// regardless of this setting.
	StructuredRunOutputEnabled *bool `jsonapi:"attr,structured-run-output-enabled,omitempty"`

	// The version of Terraform to use for this workspace.
	TerraformVersion *string `jsonapi:"attr,terraform-version,omitempty" schema:"terraform_version"`

	// List of repository-root-relative paths which list all locations to be
	// tracked for changes. See FileTriggersEnabled above for more details.
	TriggerPrefixes []string `jsonapi:"attr,trigger-prefixes,omitempty"`

	// To delete a workspace's existing VCS repo, specify null instead of an
	// object. To modify a workspace's existing VCS repo, include whichever of
	// the keys below you wish to modify. To add a new VCS repo to a workspace
	// that didn't previously have one, include at least the oauth-token-id and
	// identifier keys.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// A relative path that Terraform will execute within. This defaults to the
	// root of your repository and is typically set to a subdirectory matching
	// the environment when multiple environments exist within the same
	// repository.
	WorkingDirectory *string `jsonapi:"attr,working-directory,omitempty"`
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
	Branch            *string `json:"branch,omitempty"`
	Identifier        *string `json:"identifier,omitempty"`
	IngressSubmodules *bool   `json:"ingress-submodules,omitempty"`
	OAuthTokenID      *string `json:"oauth-token-id,omitempty"`
}
