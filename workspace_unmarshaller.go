package otf

import (
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/http/dto"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql/pggen"
)

// WorkspaceResult represents the result of a database query for a workspace.
type WorkspaceResult struct {
	WorkspaceID                pgtype.Text           `json:"workspace_id"`
	CreatedAt                  pgtype.Timestamptz    `json:"created_at"`
	UpdatedAt                  pgtype.Timestamptz    `json:"updated_at"`
	AllowDestroyPlan           bool                  `json:"allow_destroy_plan"`
	AutoApply                  bool                  `json:"auto_apply"`
	CanQueueDestroyPlan        bool                  `json:"can_queue_destroy_plan"`
	Description                pgtype.Text           `json:"description"`
	Environment                pgtype.Text           `json:"environment"`
	ExecutionMode              pgtype.Text           `json:"execution_mode"`
	FileTriggersEnabled        bool                  `json:"file_triggers_enabled"`
	GlobalRemoteState          bool                  `json:"global_remote_state"`
	MigrationEnvironment       pgtype.Text           `json:"migration_environment"`
	Name                       pgtype.Text           `json:"name"`
	QueueAllRuns               bool                  `json:"queue_all_runs"`
	SpeculativeEnabled         bool                  `json:"speculative_enabled"`
	SourceName                 pgtype.Text           `json:"source_name"`
	SourceURL                  pgtype.Text           `json:"source_url"`
	StructuredRunOutputEnabled bool                  `json:"structured_run_output_enabled"`
	TerraformVersion           pgtype.Text           `json:"terraform_version"`
	TriggerPrefixes            []string              `json:"trigger_prefixes"`
	WorkingDirectory           pgtype.Text           `json:"working_directory"`
	LockRunID                  pgtype.Text           `json:"lock_run_id"`
	LockUserID                 pgtype.Text           `json:"lock_user_id"`
	LatestRunID                pgtype.Text           `json:"latest_run_id"`
	OrganizationName           pgtype.Text           `json:"organization_name"`
	UserLock                   *pggen.Users          `json:"user_lock"`
	RunLock                    *pggen.Runs           `json:"run_lock"`
	WorkspaceRepo              *pggen.WorkspaceRepos `json:"workspace_repo"`
	Webhook                    *pggen.Webhooks       `json:"webhook"`
}

func UnmarshalWorkspaceResult(result WorkspaceResult) (*Workspace, error) {
	ws := Workspace{
		id:                         result.WorkspaceID.String,
		createdAt:                  result.CreatedAt.Time.UTC(),
		updatedAt:                  result.UpdatedAt.Time.UTC(),
		allowDestroyPlan:           result.AllowDestroyPlan,
		autoApply:                  result.AutoApply,
		canQueueDestroyPlan:        result.CanQueueDestroyPlan,
		description:                result.Description.String,
		environment:                result.Environment.String,
		executionMode:              ExecutionMode(result.ExecutionMode.String),
		fileTriggersEnabled:        result.FileTriggersEnabled,
		globalRemoteState:          result.GlobalRemoteState,
		migrationEnvironment:       result.MigrationEnvironment.String,
		name:                       result.Name.String,
		queueAllRuns:               result.QueueAllRuns,
		speculativeEnabled:         result.SpeculativeEnabled,
		structuredRunOutputEnabled: result.StructuredRunOutputEnabled,
		sourceName:                 result.SourceName.String,
		sourceURL:                  result.SourceURL.String,
		terraformVersion:           result.TerraformVersion.String,
		triggerPrefixes:            result.TriggerPrefixes,
		workingDirectory:           result.WorkingDirectory.String,
		organization:               result.OrganizationName.String,
	}

	if result.WorkspaceRepo != nil {
		ws.repo = &WorkspaceRepo{
			Branch:     result.WorkspaceRepo.Branch.String,
			ProviderID: result.WorkspaceRepo.VCSProviderID.String,
			WebhookID:  result.Webhook.WebhookID.Bytes,
			Identifier: result.Webhook.Identifier.String,
		}
	}

	if result.LatestRunID.Status == pgtype.Present {
		ws.latestRunID = &result.LatestRunID.String
	}

	if err := unmarshalWorkspaceLock(&ws, &result); err != nil {
		return nil, err
	}

	return &ws, nil
}

func MarshalWorkspaceLockParams(ws *Workspace) (pggen.UpdateWorkspaceLockByIDParams, error) {
	params := pggen.UpdateWorkspaceLockByIDParams{
		WorkspaceID: pgtype.Text{String: ws.ID(), Status: pgtype.Present},
	}
	switch lock := ws.GetLock().(type) {
	case *Unlocked:
		params.RunID = pgtype.Text{Status: pgtype.Null}
		params.UserID = pgtype.Text{Status: pgtype.Null}
	case *Run:
		params.RunID = pgtype.Text{String: lock.ID(), Status: pgtype.Present}
		params.UserID = pgtype.Text{Status: pgtype.Null}
	case *User:
		params.UserID = pgtype.Text{String: lock.ID(), Status: pgtype.Present}
		params.RunID = pgtype.Text{Status: pgtype.Null}
	default:
		return params, ErrWorkspaceInvalidLock
	}
	return params, nil
}

func unmarshalWorkspaceLock(dst *Workspace, row *WorkspaceResult) error {
	if row.UserLock == nil && row.RunLock == nil {
		dst.lock = &Unlocked{}
	} else if row.UserLock != nil {
		dst.lock = &User{id: row.UserLock.UserID.String, username: row.UserLock.Username.String}
	} else if row.RunLock != nil {
		dst.lock = &Run{id: row.RunLock.RunID.String}
	} else {
		return fmt.Errorf("workspace cannot be locked by both a run and a user")
	}
	return nil
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
		executionMode:              ExecutionMode(w.ExecutionMode),
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
		organization:               w.Organization.Name,
	}

	// The DTO only encodes whether lock is unlocked or locked, whereas our
	// domain object has three states: unlocked, run locked or user locked.
	// Therefore we ignore when DTO says lock is locked because we cannot
	// determine what/who locked it, so we assume it is unlocked.
	domain.lock = &Unlocked{}

	return &domain
}

// UnmarshalWorkspaceListJSONAPI converts a DTO into a workspace list
func UnmarshalWorkspaceListJSONAPI(json *dto.WorkspaceList) *WorkspaceList {
	wl := WorkspaceList{
		Pagination: UnmarshalPaginationJSONAPI(json.Pagination),
	}
	for _, i := range json.Items {
		wl.Items = append(wl.Items, UnmarshalWorkspaceJSONAPI(i))
	}

	return &wl
}

// WorkspacePermissionResult represents the result of a database query for a
// workspace permission.
type WorkspacePermissionResult struct {
	Role         pgtype.Text          `json:"role"`
	Team         *pggen.Teams         `json:"team"`
	Organization *pggen.Organizations `json:"organization"`
}

func UnmarshalWorkspacePermissionResult(row WorkspacePermissionResult) (*WorkspacePermission, error) {
	role, err := rbac.WorkspaceRoleFromString(row.Role.String)
	if err != nil {
		return nil, err
	}
	return &WorkspacePermission{
		Role: role,
		Team: UnmarshalTeamResult(TeamResult(*row.Team)),
	}, nil
}
