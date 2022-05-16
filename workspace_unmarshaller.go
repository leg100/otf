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
}

func UnmarshalWorkspaceListFromDB(pgresult interface{}) (workspaces []*Workspace, err error) {
	data, err := json.Marshal(pgresult)
	if err != nil {
		return nil, err
	}
	var rows []WorkspaceDBRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		ws, err := unmarshalWorkspaceDBRow(row)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}

	return workspaces, nil
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
	return unmarshalWorkspaceDBRow(row)
}

func unmarshalWorkspaceDBRow(row WorkspaceDBRow) (*Workspace, error) {
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
		SpeculativeEnabled:         row.SpeculativeEnabled,
		StructuredRunOutputEnabled: row.StructuredRunOutputEnabled,
		SourceName:                 row.SourceName,
		SourceURL:                  row.SourceUrl,
		TerraformVersion:           row.TerraformVersion,
		TriggerPrefixes:            row.TriggerPrefixes,
		WorkingDirectory:           row.WorkingDirectory,
	}

	if row.Organization != nil {
		var err error
		ws.Organization, err = UnmarshalOrganizationFromDB(row.Organization)
		if err != nil {
			return nil, err
		}
	} else if row.OrganizationID != nil {
		ws.Organization = &Organization{ID: *row.OrganizationID}
	}

	return &ws, nil
}
