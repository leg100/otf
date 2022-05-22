package dto

import "time"

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
