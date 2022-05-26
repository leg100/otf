package otf

import (
	"time"

	"github.com/leg100/otf/http/dto"
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
	UserLock                   *pggen.Users         `json:"user_lock"`
	RunLock                    *pggen.Runs          `json:"run_lock"`
	OrganizationID             string               `json:"organization_id"`
	Organization               *pggen.Organizations `json:"organization"`
}

func UnmarshalWorkspaceDBResult(row WorkspaceDBResult) (*Workspace, error) {
	ws := Workspace{
		id:                         row.WorkspaceID,
		createdAt:                  row.CreatedAt,
		updatedAt:                  row.UpdatedAt,
		allowDestroyPlan:           row.AllowDestroyPlan,
		autoApply:                  row.AutoApply,
		canQueueDestroyPlan:        row.CanQueueDestroyPlan,
		description:                row.Description,
		environment:                row.Environment,
		executionMode:              row.ExecutionMode,
		fileTriggersEnabled:        row.FileTriggersEnabled,
		globalRemoteState:          row.GlobalRemoteState,
		locked:                     row.Locked,
		migrationEnvironment:       row.MigrationEnvironment,
		name:                       row.Name,
		queueAllRuns:               row.QueueAllRuns,
		speculativeEnabled:         row.SpeculativeEnabled,
		structuredRunOutputEnabled: row.StructuredRunOutputEnabled,
		sourceName:                 row.SourceName,
		sourceURL:                  row.SourceURL,
		terraformVersion:           row.TerraformVersion,
		triggerPrefixes:            row.TriggerPrefixes,
		workingDirectory:           row.WorkingDirectory,
	}

	if row.Organization != nil {
		org, err := UnmarshalOrganizationDBResult(*row.Organization)
		if err != nil {
			return nil, err
		}
		ws.Organization = org
	} else {
		ws.Organization = &Organization{id: row.OrganizationID}
	}

	return &ws, nil
}

func UnmarshalWorkspaceDBType(typ pggen.Workspaces) (*Workspace, error) {
	ws := Workspace{
		id:                         typ.WorkspaceID,
		createdAt:                  typ.CreatedAt.Local(),
		updatedAt:                  typ.UpdatedAt.Local(),
		allowDestroyPlan:           typ.AllowDestroyPlan,
		autoApply:                  typ.AutoApply,
		canQueueDestroyPlan:        typ.CanQueueDestroyPlan,
		description:                typ.Description,
		environment:                typ.Environment,
		executionMode:              typ.ExecutionMode,
		fileTriggersEnabled:        typ.FileTriggersEnabled,
		globalRemoteState:          typ.GlobalRemoteState,
		locked:                     typ.Locked,
		migrationEnvironment:       typ.MigrationEnvironment,
		name:                       typ.Name,
		queueAllRuns:               typ.QueueAllRuns,
		speculativeEnabled:         typ.SpeculativeEnabled,
		structuredRunOutputEnabled: typ.StructuredRunOutputEnabled,
		sourceName:                 typ.SourceName,
		sourceURL:                  typ.SourceURL,
		terraformVersion:           typ.TerraformVersion,
		triggerPrefixes:            typ.TriggerPrefixes,
		workingDirectory:           typ.WorkingDirectory,
		Organization:               &Organization{id: typ.OrganizationID},
	}

	return &ws, nil
}

func UnmarshalWorkspaceJSONAPI(w *dto.Workspace) *Workspace {
	domain := Workspace{
		id:                         w.ID,
		allowDestroyPlan:           w.AllowDestroyPlan,
		autoApply:                  w.AutoApply,
		canQueueDestroyPlan:        w.CanQueueDestroyPlan,
		createdAt:                  w.CreatedAt,
		updatedAt:                  w.UpdatedAt,
		description:                w.Description,
		environment:                w.Environment,
		executionMode:              w.ExecutionMode,
		fileTriggersEnabled:        w.FileTriggersEnabled,
		globalRemoteState:          w.GlobalRemoteState,
		locked:                     w.Locked,
		migrationEnvironment:       w.MigrationEnvironment,
		name:                       w.Name,
		queueAllRuns:               w.QueueAllRuns,
		speculativeEnabled:         w.SpeculativeEnabled,
		sourceName:                 w.SourceName,
		sourceURL:                  w.SourceURL,
		structuredRunOutputEnabled: w.StructuredRunOutputEnabled,
		terraformVersion:           w.TerraformVersion,
		workingDirectory:           w.WorkingDirectory,
		triggerPrefixes:            w.TriggerPrefixes,
	}

	if w.Organization != nil {
		domain.Organization = UnmarshalOrganizationJSONAPI(w.Organization)
	}

	return &domain
}

func UnmarshalWorkspaceListJSONAPI(json *dto.WorkspaceList) *WorkspaceList {
	pagination := Pagination(*json.Pagination)
	wl := WorkspaceList{
		Pagination: &pagination,
	}
	for _, i := range json.Items {
		wl.Items = append(wl.Items, UnmarshalWorkspaceJSONAPI(i))
	}

	return &wl
}
