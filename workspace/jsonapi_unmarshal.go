package workspace

import (
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func unmarshalJSONAPI(w *jsonapi.Workspace) *Workspace {
	domain := Workspace{
		id:                         w.ID,
		allowDestroyPlan:           w.AllowDestroyPlan,
		autoApply:                  w.AutoApply,
		canQueueDestroyPlan:        w.CanQueueDestroyPlan,
		createdAt:                  w.CreatedAt,
		updatedAt:                  w.UpdatedAt,
		description:                w.Description,
		environment:                w.Environment,
		executionMode:              otf.ExecutionMode(w.ExecutionMode),
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
	domain.state = nil

	return &domain
}

// unmarshalListJSONAPI converts a DTO into a workspace list
func unmarshalListJSONAPI(json *jsonapi.WorkspaceList) *WorkspaceList {
	wl := WorkspaceList{
		Pagination: otf.NewPaginationFromJSONAPI(json.Pagination),
	}
	for _, i := range json.Items {
		wl.Items = append(wl.Items, unmarshalJSONAPI(i))
	}

	return &wl
}
