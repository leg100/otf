package otf

import (
	"encoding/json"
	"time"
)

type WorkspaceDBRow struct {
	WorkspaceID                string             `json:"workspace_id"`
	CreatedAt                  time.Time          `json:"created_at"`
	UpdatedAt                  time.Time          `json:"updated_at"`
	AllowDestroyPlan           bool               `json:"allow_destroy_plan"`
	AutoApply                  bool               `json:"auto_apply"`
	CanQueueDestroyPlan        bool               `json:"can_queue_destroy_plan"`
	Description                string             `json:"description"`
	Environment                string             `json:"environment"`
	ExecutionMode              string             `json:"execution_mode"`
	FileTriggersEnabled        bool               `json:"file_triggers_enabled"`
	GlobalRemoteState          bool               `json:"global_remote_state"`
	Locked                     bool               `json:"locked"`
	MigrationEnvironment       string             `json:"migration_environment"`
	Name                       string             `json:"name"`
	QueueAllRuns               bool               `json:"queue_all_runs"`
	SpeculativeEnabled         bool               `json:"speculative_enabled"`
	SourceName                 string             `json:"source_name"`
	SourceUrl                  string             `json:"source_url"`
	StructuredRunOutputEnabled bool               `json:"structured_run_output_enabled"`
	TerraformVersion           string             `json:"terraform_version"`
	TriggerPrefixes            []string           `json:"trigger_prefixes"`
	WorkingDirectory           string             `json:"working_directory"`
	OrganizationID             *string            `json:"organization_id"`
	Organization               *OrganizationDBRow `json:"organization"`
	FullCount                  int                `json:"full_count"`
}

func UnmarshalWorkspaceFromDB(pgresult interface{}) (*Workspace, error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var row WorkspaceDBRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}

	ws := Workspace{
		ID:               row.WorkspaceID,
		Name:             row.Name,
		WorkingDirectory: row.WorkingDirectory,
		TerraformVersion: row.TerraformVersion,
		Timestamps: Timestamps{
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
	}

	if row.Organization != nil {
		ws.Organization, err = UnmarshalOrganizationFromDB(row.Organization)
		if err != nil {
			return nil, err
		}
	}
	if row.OrganizationID != nil {
		ws.Organization = &Organization{ID: *row.OrganizationID}
	}

	return &ws, nil
}
