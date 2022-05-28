package otf

import (
	"fmt"
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
	LockRunID                  string               `json:"lock_run_id"`
	LockUserID                 string               `json:"lock_user_id"`
	UserLock                   *pggen.Users         `json:"user_lock"`
	RunLock                    *pggen.Runs          `json:"run_lock"`
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

	if row.UserLock == nil && row.RunLock == nil {
		ws.lock = &Unlocked{}
	} else if row.UserLock != nil {
		ws.lock = &User{id: row.UserLock.UserID, username: row.UserLock.Username}
	} else if row.RunLock != nil {
		ws.lock = &Run{id: row.RunLock.RunID}
	} else {
		return nil, fmt.Errorf("workspace cannot be locked by both a run and a user")
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
		id:                   typ.WorkspaceID,
		createdAt:            typ.CreatedAt.Local(),
		updatedAt:            typ.UpdatedAt.Local(),
		allowDestroyPlan:     typ.AllowDestroyPlan,
		autoApply:            typ.AutoApply,
		canQueueDestroyPlan:  typ.CanQueueDestroyPlan,
		description:          typ.Description,
		environment:          typ.Environment,
		executionMode:        typ.ExecutionMode,
		fileTriggersEnabled:  typ.FileTriggersEnabled,
		globalRemoteState:    typ.GlobalRemoteState,
		migrationEnvironment: typ.MigrationEnvironment,
		// Assume workspace is unlocked
		lock:                       &Unlocked{},
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

	// The DTO only encodes whether lock is unlocked or locked, whereas our
	// domain object has three states: unlocked, run locked or user locked.
	// Therefore we ignore when DTO says lock is locked because we cannot
	// determine what/who locked it, so we can assume it is unlocked.
	domain.lock = &Unlocked{}

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
