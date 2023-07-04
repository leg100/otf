package workspace

import (
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/resource"
)

func unmarshalJSONAPI(w *types.Workspace) *Workspace {
	domain := Workspace{
		ID:                         w.ID,
		AllowDestroyPlan:           w.AllowDestroyPlan,
		AutoApply:                  w.AutoApply,
		CanQueueDestroyPlan:        w.CanQueueDestroyPlan,
		CreatedAt:                  w.CreatedAt,
		UpdatedAt:                  w.UpdatedAt,
		Description:                w.Description,
		Environment:                w.Environment,
		ExecutionMode:              ExecutionMode(w.ExecutionMode),
		FileTriggersEnabled:        w.FileTriggersEnabled,
		GlobalRemoteState:          w.GlobalRemoteState,
		MigrationEnvironment:       w.MigrationEnvironment,
		Name:                       w.Name,
		QueueAllRuns:               w.QueueAllRuns,
		SpeculativeEnabled:         w.SpeculativeEnabled,
		SourceName:                 w.SourceName,
		SourceURL:                  w.SourceURL,
		StructuredRunOutputEnabled: w.StructuredRunOutputEnabled,
		TerraformVersion:           w.TerraformVersion,
		WorkingDirectory:           w.WorkingDirectory,
		TriggerPrefixes:            w.TriggerPrefixes,
		Organization:               w.Organization.Name,
	}

	// The DTO only encodes whether lock is unlocked or locked, whereas our
	// domain object has three states: unlocked, run locked or user locked.
	// Therefore we ignore when DTO says lock is locked because we cannot
	// determine what/who locked it, so we assume it is unlocked.
	domain.Lock = nil

	return &domain
}

// unmarshalListJSONAPI converts a DTO into a workspace list
func unmarshalListJSONAPI(json *types.WorkspaceList) *resource.Page[*Workspace] {
	wl := resource.Page[*Workspace]{
		Pagination: (*resource.Pagination)(json.Pagination),
	}
	for _, i := range json.Items {
		wl.Items = append(wl.Items, unmarshalJSONAPI(i))
	}
	return &wl
}
