package otf

import (
	"time"

	"github.com/leg100/otf/sql/pggen"
)

type WorkspaceDBResult struct {
	WorkspaceID                string               `json:"workspace_id"`
	CreatedAt                  time.Time            `json:"created_at"`
	UpdatedAt                  time.Time            `json:"updated_at"`
	AllowDestroyPlan           bool                 `json:"allow_destroy_plan"`
	AutoApply                  bool                 `json:"auto_apply"`
	CanQueueDestroyPlan        bool                 `json:"can_queue_destroy_plan"`
	Description                string               `json:"description"`
	Environment                string               `json:"environment"`
	ExecutionMode              string               `json:"execution_mode"`
	FileTriggersEnabled        bool                 `json:"file_triggers_enabled"`
	GlobalRemoteState          bool                 `json:"global_remote_state"`
	Locked                     bool                 `json:"locked"`
	MigrationEnvironment       string               `json:"migration_environment"`
	Name                       string               `json:"name"`
	QueueAllRuns               bool                 `json:"queue_all_runs"`
	SpeculativeEnabled         bool                 `json:"speculative_enabled"`
	SourceName                 string               `json:"source_name"`
	SourceURL                  string               `json:"source_url"`
	StructuredRunOutputEnabled bool                 `json:"structured_run_output_enabled"`
	TerraformVersion           string               `json:"terraform_version"`
	TriggerPrefixes            []string             `json:"trigger_prefixes"`
	WorkingDirectory           string               `json:"working_directory"`
	OrganizationID             string               `json:"organization_id"`
	Organization               *pggen.Organizations `json:"organization"`
}

func UnmarshalWorkspaceDBResult(row WorkspaceDBResult) (*Workspace, error) {
	ws := Workspace{
		ID: row.WorkspaceID,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		AllowDestroyPlan:           row.AllowDestroyPlan,
		AutoApply:                  row.AutoApply,
		CanQueueDestroyPlan:        row.CanQueueDestroyPlan,
		Description:                row.Description,
		Environment:                row.Environment,
		ExecutionMode:              row.ExecutionMode,
		FileTriggersEnabled:        row.FileTriggersEnabled,
		GlobalRemoteState:          row.GlobalRemoteState,
		Locked:                     row.Locked,
		MigrationEnvironment:       row.MigrationEnvironment,
		Name:                       row.Name,
		QueueAllRuns:               row.QueueAllRuns,
		SpeculativeEnabled:         row.SpeculativeEnabled,
		StructuredRunOutputEnabled: row.StructuredRunOutputEnabled,
		SourceName:                 row.SourceName,
		SourceURL:                  row.SourceURL,
		TerraformVersion:           row.TerraformVersion,
		TriggerPrefixes:            row.TriggerPrefixes,
		WorkingDirectory:           row.WorkingDirectory,
	}

	org, err := UnmarshalOrganizationDBResult(*row.Organization)
	if err != nil {
		return nil, err
	}
	ws.Organization = org

	return &ws, nil
}

func unmarshalWorkspaceDBType(typ *pggen.Workspaces) (*Workspace, error) {
	ws := Workspace{
		ID: typ.WorkspaceID,
		Timestamps: Timestamps{
			CreatedAt: typ.CreatedAt.Local(),
			UpdatedAt: typ.UpdatedAt.Local(),
		},
		AllowDestroyPlan:           typ.AllowDestroyPlan,
		AutoApply:                  typ.AutoApply,
		CanQueueDestroyPlan:        typ.CanQueueDestroyPlan,
		Description:                typ.Description,
		Environment:                typ.Environment,
		ExecutionMode:              typ.ExecutionMode,
		FileTriggersEnabled:        typ.FileTriggersEnabled,
		GlobalRemoteState:          typ.GlobalRemoteState,
		Locked:                     typ.Locked,
		MigrationEnvironment:       typ.MigrationEnvironment,
		Name:                       typ.Name,
		QueueAllRuns:               typ.QueueAllRuns,
		SpeculativeEnabled:         typ.SpeculativeEnabled,
		StructuredRunOutputEnabled: typ.StructuredRunOutputEnabled,
		SourceName:                 typ.SourceName,
		SourceURL:                  typ.SourceURL,
		TerraformVersion:           typ.TerraformVersion,
		TriggerPrefixes:            typ.TriggerPrefixes,
		WorkingDirectory:           typ.WorkingDirectory,
		Organization:               &Organization{ID: typ.OrganizationID},
	}

	return &ws, nil
}
