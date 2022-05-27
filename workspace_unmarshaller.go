package otf

import (
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/sql/pggen"
)

type WorkspaceDBResult struct {
	WorkspaceID                pgtype.Text          `json:"workspace_id"`
	CreatedAt                  time.Time            `json:"created_at"`
	UpdatedAt                  time.Time            `json:"updated_at"`
	AllowDestroyPlan           bool                 `json:"allow_destroy_plan"`
	AutoApply                  bool                 `json:"auto_apply"`
	CanQueueDestroyPlan        bool                 `json:"can_queue_destroy_plan"`
	Description                pgtype.Text          `json:"description"`
	Environment                pgtype.Text          `json:"environment"`
	ExecutionMode              pgtype.Text          `json:"execution_mode"`
	FileTriggersEnabled        bool                 `json:"file_triggers_enabled"`
	GlobalRemoteState          bool                 `json:"global_remote_state"`
	Locked                     bool                 `json:"locked"`
	MigrationEnvironment       pgtype.Text          `json:"migration_environment"`
	Name                       pgtype.Text          `json:"name"`
	QueueAllRuns               bool                 `json:"queue_all_runs"`
	SpeculativeEnabled         bool                 `json:"speculative_enabled"`
	SourceName                 pgtype.Text          `json:"source_name"`
	SourceURL                  pgtype.Text          `json:"source_url"`
	StructuredRunOutputEnabled bool                 `json:"structured_run_output_enabled"`
	TerraformVersion           pgtype.Text          `json:"terraform_version"`
	TriggerPrefixes            []string             `json:"trigger_prefixes"`
	WorkingDirectory           pgtype.Text          `json:"working_directory"`
	OrganizationID             pgtype.Text          `json:"organization_id"`
	Organization               *pggen.Organizations `json:"organization"`
}

func UnmarshalWorkspaceDBResult(row WorkspaceDBResult) (*Workspace, error) {
	ws := Workspace{
		id:                         row.WorkspaceID.String,
		createdAt:                  row.CreatedAt,
		updatedAt:                  row.UpdatedAt,
		allowDestroyPlan:           row.AllowDestroyPlan,
		autoApply:                  row.AutoApply,
		canQueueDestroyPlan:        row.CanQueueDestroyPlan,
		description:                row.Description.String,
		environment:                row.Environment.String,
		executionMode:              row.ExecutionMode.String,
		fileTriggersEnabled:        row.FileTriggersEnabled,
		globalRemoteState:          row.GlobalRemoteState,
		locked:                     row.Locked,
		migrationEnvironment:       row.MigrationEnvironment.String,
		name:                       row.Name.String,
		queueAllRuns:               row.QueueAllRuns,
		speculativeEnabled:         row.SpeculativeEnabled,
		structuredRunOutputEnabled: row.StructuredRunOutputEnabled,
		sourceName:                 row.SourceName.String,
		sourceURL:                  row.SourceURL.String,
		terraformVersion:           row.TerraformVersion.String,
		triggerPrefixes:            row.TriggerPrefixes,
		workingDirectory:           row.WorkingDirectory.String,
	}

	if row.Organization != nil {
		org, err := UnmarshalOrganizationDBResult(*row.Organization)
		if err != nil {
			return nil, err
		}
		ws.Organization = org
	} else {
		ws.Organization = &Organization{id: row.OrganizationID.String}
	}

	return &ws, nil
}

func UnmarshalWorkspaceDBType(typ pggen.Workspaces) (*Workspace, error) {
	ws := Workspace{
		id:                         typ.WorkspaceID.String,
		createdAt:                  typ.CreatedAt.Local(),
		updatedAt:                  typ.UpdatedAt.Local(),
		allowDestroyPlan:           typ.AllowDestroyPlan,
		autoApply:                  typ.AutoApply,
		canQueueDestroyPlan:        typ.CanQueueDestroyPlan,
		description:                typ.Description.String,
		environment:                typ.Environment.String,
		executionMode:              typ.ExecutionMode.String,
		fileTriggersEnabled:        typ.FileTriggersEnabled,
		globalRemoteState:          typ.GlobalRemoteState,
		locked:                     typ.Locked,
		migrationEnvironment:       typ.MigrationEnvironment.String,
		name:                       typ.Name.String,
		queueAllRuns:               typ.QueueAllRuns,
		speculativeEnabled:         typ.SpeculativeEnabled,
		structuredRunOutputEnabled: typ.StructuredRunOutputEnabled,
		sourceName:                 typ.SourceName.String,
		sourceURL:                  typ.SourceURL.String,
		terraformVersion:           typ.TerraformVersion.String,
		triggerPrefixes:            typ.TriggerPrefixes,
		workingDirectory:           typ.WorkingDirectory.String,
		Organization:               &Organization{id: typ.OrganizationID.String},
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
