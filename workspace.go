package otf

import (
	"context"
	"errors"
	"fmt"
)

const (
	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true
	DefaultTerraformVersion    = "1.0.10"
	DefaultExecutionMode       = "remote"
)

var (
	ErrWorkspaceAlreadyLocked   = errors.New("workspace already locked")
	ErrWorkspaceAlreadyUnlocked = errors.New("workspace already unlocked")
	ErrInvalidWorkspaceSpec     = errors.New("invalid workspace spec options")
)

// Workspace represents a Terraform Enterprise workspace.
type Workspace struct {
	ID string `db:"workspace_id" jsonapi:"primary,workspaces" schema:"workspace_id"`

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	AllowDestroyPlan           bool
	AutoApply                  bool
	CanQueueDestroyPlan        bool
	Description                string
	Environment                string
	ExecutionMode              string
	FileTriggersEnabled        bool
	GlobalRemoteState          bool
	Locked                     bool
	MigrationEnvironment       string
	Name                       string `schema:"workspace_name"`
	QueueAllRuns               bool
	SpeculativeEnabled         bool
	StructuredRunOutputEnabled bool
	SourceName                 string
	SourceURL                  string `db:"source_url"`
	TerraformVersion           string
	TriggerPrefixes            CSV
	VCSRepo                    *VCSRepo
	WorkingDirectory           string

	// Workspace belongs to an organization
	Organization *Organization `db:"organizations"`
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
	Organization string `schema:"organization_name"`

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

// VCSRepoOptions is used by workspaces, policy sets, and registry modules
// VCSRepoOptions represents the configuration options of a VCS integration.
type VCSRepoOptions struct {
	Branch            *string `json:"branch,omitempty"`
	Identifier        *string `json:"identifier,omitempty"`
	IngressSubmodules *bool   `json:"ingress-submodules,omitempty"`
	OAuthTokenID      *string `json:"oauth-token-id,omitempty"`
}

// WorkspaceLockOptions represents the options for locking a workspace.
type WorkspaceLockOptions struct {
	// Specifies the reason for locking the workspace.
	Reason *string `jsonapi:"attr,reason,omitempty"`
}

// WorkspaceList represents a list of Workspaces.
type WorkspaceList struct {
	*Pagination
	Items []*Workspace
}

type WorkspaceService interface {
	Create(ctx context.Context, opts WorkspaceCreateOptions) (*Workspace, error)
	Get(ctx context.Context, spec WorkspaceSpec) (*Workspace, error)
	List(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error)
	Update(ctx context.Context, spec WorkspaceSpec, opts WorkspaceUpdateOptions) (*Workspace, error)
	Lock(ctx context.Context, spec WorkspaceSpec, opts WorkspaceLockOptions) (*Workspace, error)
	Unlock(ctx context.Context, spec WorkspaceSpec) (*Workspace, error)
	Delete(ctx context.Context, spec WorkspaceSpec) error
}

type WorkspaceStore interface {
	Create(ws *Workspace) (*Workspace, error)
	Get(spec WorkspaceSpec) (*Workspace, error)
	List(opts WorkspaceListOptions) (*WorkspaceList, error)
	Update(spec WorkspaceSpec, fn func(ws *Workspace, updater WorkspaceUpdater) error) (*Workspace, error)
	Delete(spec WorkspaceSpec) error
}

type WorkspaceUpdater interface {
	ToggleLock(ctx context.Context, lock bool) error
	UpdateName(ctx context.Context, name string) error
	UpdateAllowDestroyPlan(ctx context.Context, allow bool) error
}

// WorkspaceSpec is used for identifying an individual workspace. Either ID
// *or* both Name and OrganizationName must be specfiied.
type WorkspaceSpec struct {
	// Specify workspace using its ID
	ID *string `db:"workspace_id"`

	// Specify workspace using its name and organization
	Name             *string `schema:"workspace_name"`
	OrganizationName *string `schema:"organization_name"`
}

// WorkspaceListOptions are options for paginating and filtering a list of
// Workspaces
type WorkspaceListOptions struct {
	// Pagination
	ListOptions

	// Optionally filter workspaces with name matching prefix
	Prefix *string `schema:"search[name],omitempty"`

	// OrganizationName filters workspaces by organization name
	OrganizationName *string `schema:"organization_name,omitempty"`

	// A list of relations to include. See available resources https://www.terraform.io/docs/cloud/api/workspaces.html#available-related-resources
	Include *string `schema:"include"`
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

func (ws *Workspace) GetID() string  { return ws.ID }
func (ws *Workspace) String() string { return ws.ID }

func NewWorkspace(opts WorkspaceCreateOptions, org *Organization) *Workspace {
	ws := Workspace{
		ID:                  NewID("ws"),
		Timestamps:          NewTimestamps(),
		Name:                *opts.Name,
		AllowDestroyPlan:    DefaultAllowDestroyPlan,
		ExecutionMode:       DefaultExecutionMode,
		FileTriggersEnabled: DefaultFileTriggersEnabled,
		GlobalRemoteState:   true, // Only global remote state is supported
		TerraformVersion:    DefaultTerraformVersion,
		SpeculativeEnabled:  true,
		Organization:        org,
	}

	// TODO: ExecutionMode and Operations are mututally exclusive options, this
	// should be enforced.
	if opts.ExecutionMode != nil {
		ws.ExecutionMode = *opts.ExecutionMode
	}
	// Operations is deprecated in favour of ExecutionMode.
	if opts.Operations != nil {
		if *opts.Operations {
			ws.ExecutionMode = "remote"
		} else {
			ws.ExecutionMode = "local"
		}
	}

	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SourceName != nil {
		ws.SourceName = *opts.SourceName
	}
	if opts.SourceURL != nil {
		ws.SourceURL = *opts.SourceURL
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		ws.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}

	return &ws
}

func (o WorkspaceCreateOptions) Valid() error {
	if !validString(o.Name) {
		return ErrRequiredName
	}
	if !ValidStringID(o.Name) {
		return ErrInvalidName
	}
	if o.TerraformVersion != nil && !validSemanticVersion(*o.TerraformVersion) {
		return ErrInvalidTerraformVersion
	}
	if o.Operations != nil && o.ExecutionMode != nil {
		return errors.New("operations is deprecated and cannot be specified when execution mode is used")
	}
	if o.AgentPoolID != nil && (o.ExecutionMode == nil || *o.ExecutionMode != "agent") {
		return errors.New("specifying an agent pool ID requires 'agent' execution mode")
	}
	if o.AgentPoolID == nil && (o.ExecutionMode != nil && *o.ExecutionMode == "agent") {
		return errors.New("'agent' execution mode requires an agent pool ID to be specified")
	}

	return nil
}

func (o WorkspaceUpdateOptions) Valid() error {
	if o.Name != nil && !ValidStringID(o.Name) {
		return ErrInvalidName
	}
	if o.TerraformVersion != nil && !validSemanticVersion(*o.TerraformVersion) {
		return ErrInvalidTerraformVersion
	}
	if o.Operations != nil && o.ExecutionMode != nil {
		return errors.New("operations is deprecated and cannot be specified when execution mode is used")
	}
	if o.AgentPoolID == nil && (o.ExecutionMode != nil && *o.ExecutionMode == "agent") {
		return errors.New("'agent' execution mode requires an agent pool ID to be specified")
	}

	return nil
}

// ToggleLock toggles the workspace lock.
func (ws *Workspace) ToggleLock(lock bool, updater WorkspaceUpdater) error {
	if lock && ws.Locked {
		return ErrWorkspaceAlreadyLocked
	}
	if !lock && !ws.Locked {
		return ErrWorkspaceAlreadyUnlocked
	}

	ws.Locked = lock

	return updater.ToggleLock(context.Background(), lock)
}

func (ws *Workspace) UpdateWithOptions(ctx context.Context, opts WorkspaceUpdateOptions, updater WorkspaceUpdater) error {
	if opts.Name != nil {
		if err := updater.UpdateName(ctx, *opts.Name); err != nil {
			return err
		}
	}
	if opts.AllowDestroyPlan != nil {
		if err := updater.UpdateAllowDestroyPlan(ctx, *opts.AllowDestroyPlan); err != nil {
			return err
		}
	}

	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.ExecutionMode != nil {
		ws.ExecutionMode = *opts.ExecutionMode
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.Operations != nil {
		if *opts.Operations {
			ws.ExecutionMode = "remote"
		} else {
			ws.ExecutionMode = "local"
		}
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		ws.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = opts.TriggerPrefixes
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}

	return nil
}

func (spec *WorkspaceSpec) String() string {
	switch {
	case spec.ID != nil:
		return *spec.ID
	case spec.Name != nil && spec.OrganizationName != nil:
		return *spec.OrganizationName + "/" + *spec.Name
	default:
		panic("invalid workspace spec")
	}
}

func (spec *WorkspaceSpec) Valid() error {
	if spec.ID != nil {
		if *spec.ID == "" {
			return fmt.Errorf("id is an empty string")
		}
		return nil
	}

	// No ID specified; both org and workspace name must be specified

	if spec.Name == nil {
		return fmt.Errorf("workspace name nor id specified")
	}

	if spec.OrganizationName == nil {
		return fmt.Errorf("must specify both organization and workspace")
	}

	if *spec.Name == "" {
		return fmt.Errorf("workspace name is an empty string")
	}

	if *spec.OrganizationName == "" {
		return fmt.Errorf("organization name is an empty string")
	}

	return nil
}
