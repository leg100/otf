package trigger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
)

type TFEAPI struct {
	*tfeapi.Responder
	Client tfeClient
}

type tfeClient interface {
	CreateRunTrigger(ctx context.Context, workspaceID, sourceableWorkspaceID resource.TfeID) (*trigger, error)
	ListRunTriggers(ctx context.Context, opts ListOptions) ([]*trigger, error)
	GetRunTrigger(ctx context.Context, triggerID resource.TfeID) (*trigger, error)
	DeleteRunTrigger(ctx context.Context, triggerID resource.TfeID) error
}

// TFERunTrigger represents a run trigger in the TFE API.
type TFERunTrigger struct {
	ID         resource.TfeID          `jsonapi:"primary,run-triggers"`
	CreatedAt  time.Time               `jsonapi:"attribute" json:"created-at"`
	Workspace  *workspace.TFEWorkspace `jsonapi:"relationship" json:"workspace"`
	Sourceable *workspace.TFEWorkspace `jsonapi:"relationship" json:"sourceable"`
}

func (a *TFEAPI) AddHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/run-triggers", a.createRunTrigger).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/run-triggers", a.listRunTriggers).Methods("GET")
	r.HandleFunc("/run-triggers/{id}", a.getRunTrigger).Methods("GET")
	r.HandleFunc("/run-triggers/{id}", a.deleteRunTrigger).Methods("DELETE")
}

func (a *TFEAPI) createRunTrigger(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params struct {
		// Type is used by JSON:API to set the resource type.
		Type string `jsonapi:"primary,run-triggers"`

		// The source workspace that triggers runs in the target workspace.
		Sourceable *workspace.TFEWorkspace `jsonapi:"relationship" json:"sourceable"`
	}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if params.Sourceable == nil {
		tfeapi.Error(w, fmt.Errorf("missing required relationship: sourceable"))
		return
	}

	t, err := a.Client.CreateRunTrigger(r.Context(), workspaceID, params.Sourceable.ID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(t), http.StatusCreated)
}

func (a *TFEAPI) listRunTriggers(w http.ResponseWriter, r *http.Request) {
	var opts ListOptions
	if err := decode.All(&opts, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	triggers, err := a.Client.ListRunTriggers(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*TFERunTrigger, len(triggers))
	for i, t := range triggers {
		to[i] = a.convert(t)
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *TFEAPI) getRunTrigger(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	t, err := a.Client.GetRunTrigger(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convert(t), http.StatusOK)
}

func (a *TFEAPI) deleteRunTrigger(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Client.DeleteRunTrigger(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) convert(from *trigger) *TFERunTrigger {
	return &TFERunTrigger{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		Workspace: &workspace.TFEWorkspace{
			ID: from.WorkspaceID,
		},
		Sourceable: &workspace.TFEWorkspace{
			ID: from.SourceableWorkspaceID,
		},
	}
}
