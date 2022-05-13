package sql

import (
	"time"

	"github.com/leg100/otf"
)

type workspaceRow struct {
	WorkspaceID                *string       `json:"workspace_id"`
	CreatedAt                  time.Time     `json:"created_at"`
	UpdatedAt                  time.Time     `json:"updated_at"`
	AllowDestroyPlan           *bool         `json:"allow_destroy_plan"`
	AutoApply                  *bool         `json:"auto_apply"`
	CanQueueDestroyPlan        *bool         `json:"can_queue_destroy_plan"`
	Description                *string       `json:"description"`
	Environment                *string       `json:"environment"`
	ExecutionMode              *string       `json:"execution_mode"`
	FileTriggersEnabled        *bool         `json:"file_triggers_enabled"`
	GlobalRemoteState          *bool         `json:"global_remote_state"`
	Locked                     *bool         `json:"locked"`
	MigrationEnvironment       *string       `json:"migration_environment"`
	Name                       *string       `json:"name"`
	QueueAllRuns               *bool         `json:"queue_all_runs"`
	SpeculativeEnabled         *bool         `json:"speculative_enabled"`
	SourceName                 *string       `json:"source_name"`
	SourceUrl                  *string       `json:"source_url"`
	StructuredRunOutputEnabled *bool         `json:"structured_run_output_enabled"`
	TerraformVersion           *string       `json:"terraform_version"`
	TriggerPrefixes            []string      `json:"trigger_prefixes"`
	WorkingDirectory           *string       `json:"working_directory"`
	OrganizationID             *string       `json:"organization_id"`
	Organization               Organizations `json:"organization"`
}

func (row workspaceRow) convert() *otf.Workspace {
	ws := otf.Workspace{}
	ws.ID = *row.WorkspaceID
	ws.CreatedAt = row.CreatedAt
	ws.UpdatedAt = row.UpdatedAt
	ws.AllowDestroyPlan = *row.AllowDestroyPlan
	ws.AutoApply = *row.AutoApply
	ws.CanQueueDestroyPlan = *row.CanQueueDestroyPlan
	ws.Description = *row.Description
	ws.Environment = *row.Environment
	ws.ExecutionMode = *row.ExecutionMode
	ws.FileTriggersEnabled = *row.FileTriggersEnabled
	ws.GlobalRemoteState = *row.GlobalRemoteState
	ws.Locked = *row.Locked
	ws.MigrationEnvironment = *row.MigrationEnvironment
	ws.Name = *row.Name
	ws.QueueAllRuns = *row.QueueAllRuns
	ws.SpeculativeEnabled = *row.SpeculativeEnabled
	ws.StructuredRunOutputEnabled = *row.StructuredRunOutputEnabled
	ws.SourceName = *row.SourceName
	ws.SourceURL = *row.SourceUrl
	ws.TerraformVersion = *row.TerraformVersion
	ws.TriggerPrefixes = row.TriggerPrefixes
	ws.WorkingDirectory = *row.WorkingDirectory

	ws.Organization = convertOrganizationComposite(Organizations(row.Organization))
	return &ws
}

func convertWorkspaceComposite(row Workspaces) *otf.Workspace {
	ws := otf.Workspace{}
	ws.ID = *row.WorkspaceID
	ws.CreatedAt = row.CreatedAt
	ws.UpdatedAt = row.UpdatedAt
	ws.AllowDestroyPlan = *row.AllowDestroyPlan
	ws.AutoApply = *row.AutoApply
	ws.CanQueueDestroyPlan = *row.CanQueueDestroyPlan
	ws.Description = *row.Description
	ws.Environment = *row.Environment
	ws.ExecutionMode = *row.ExecutionMode
	ws.FileTriggersEnabled = *row.FileTriggersEnabled
	ws.GlobalRemoteState = *row.GlobalRemoteState
	ws.Locked = *row.Locked
	ws.MigrationEnvironment = *row.MigrationEnvironment
	ws.Name = *row.Name
	ws.QueueAllRuns = *row.QueueAllRuns
	ws.SpeculativeEnabled = *row.SpeculativeEnabled
	ws.StructuredRunOutputEnabled = *row.StructuredRunOutputEnabled
	ws.SourceName = *row.SourceName
	ws.SourceURL = *row.SourceUrl
	ws.TerraformVersion = *row.TerraformVersion
	ws.TriggerPrefixes = row.TriggerPrefixes
	ws.WorkingDirectory = *row.WorkingDirectory

	return &ws
}
