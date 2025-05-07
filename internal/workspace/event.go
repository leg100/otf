package workspace

import (
	"time"

	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
)

type Event struct {
	ID                         resource.TfeID    `json:"workspace_id"`
	CreatedAt                  time.Time         `json:"created_at"`
	UpdatedAt                  time.Time         `json:"updated_at"`
	AllowDestroyPlan           bool              `json:"allow_destroy_plan"`
	AutoApply                  bool              `json:"auto_apply"`
	CanQueueDestroyPlan        bool              `json:"can_queue_destroy_plan"`
	Description                string            `json:"description"`
	Environment                string            `json:"environment"`
	ExecutionMode              ExecutionMode     `json:"execution_mode"`
	GlobalRemoteState          bool              `json:"global_remote_state"`
	MigrationEnvironment       string            `json:"migration_environment"`
	Name                       string            `json:"name"`
	QueueAllRuns               bool              `json:"queue_all_runs"`
	SpeculativeEnabled         bool              `json:"speculative_enabled"`
	SourceName                 string            `json:"source_name"`
	SourceURL                  string            `json:"source_url"`
	StructuredRunOutputEnabled bool              `json:"structured_run_output_enabled"`
	EngineVersion              *Version          `json:"engine_version"`
	TriggerPrefixes            []string          `json:"trigger_prefixes"`
	WorkingDirectory           string            `json:"working_directory"`
	LockRunID                  *resource.TfeID   `json:"lock_run_id"`
	LatestRunID                *resource.TfeID   `json:"latest_run_id"`
	Organization               organization.Name `json:"organization_name"`
	Branch                     string            `json:"branch"`
	CurrentStateVersionID      *resource.TfeID   `json:"current_state_version_id"`
	TriggerPatterns            []string          `json:"trigger_patterns"`
	VCSTagsRegex               *string           `json:"vcs_tags_regex"`
	AllowCLIApply              bool              `json:"allow_cli_apply"`
	AgentPoolID                *resource.TfeID   `json:"agent_pool_id"`
	LockUsername               *user.Username    `json:"lock_username"`
	Engine                     *engine.Engine    `json:"engine"`
}
