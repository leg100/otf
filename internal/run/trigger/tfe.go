package trigger

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
)

const IncludeSourceable tfeapi.IncludeName = "sourceable"

type TFEAPI struct {
	*tfeapi.Responder
	Client     tfeClient
	Authorizer *authz.Authorizer
}

type tfeClient interface {
	CreateRunTrigger(ctx context.Context, workspaceID, sourceableWorkspaceID resource.TfeID) (*trigger, error)
	ListRunTriggers(ctx context.Context, opts ListOptions) ([]*trigger, error)
	GetRunTrigger(ctx context.Context, triggerID resource.TfeID) (*trigger, error)
	DeleteRunTrigger(ctx context.Context, triggerID resource.TfeID) error
	GetWorkspace(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
}

// TFERunTrigger represents a run trigger in the TFE API.
type TFERunTrigger struct {
	ID         resource.TfeID          `jsonapi:"primary,run-triggers"`
	CreatedAt  time.Time               `jsonapi:"attribute" json:"created-at"`
	Workspace  *workspace.TFEWorkspace `jsonapi:"relationship" json:"workspace"`
	Sourceable *workspace.TFEWorkspace `jsonapi:"relationship" json:"sourceable"`
}

func NewTFEAPI(
	client tfeClient,
	authorizer *authz.Authorizer,
	responder *tfeapi.Responder,
) *TFEAPI {
	api := &TFEAPI{
		Client:     client,
		Authorizer: authorizer,
		Responder:  responder,
	}
	responder.Register(tfeapi.IncludeWorkspace, api.includeWorkspace)
	responder.Register(IncludeSourceable, api.includeSourceable)
	return api
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
		if errors.Is(err, errTriggerLoop) {
			tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		} else {
			tfeapi.Error(w, err)
		}
		return
	}

	a.Respond(w, r, a.convert(t), http.StatusCreated)
}

func (a *TFEAPI) listRunTriggers(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		types.PageOptions
		ListOptions
	}
	if err := decode.All(&opts, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	triggers, err := a.Client.ListRunTriggers(r.Context(), opts.ListOptions)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	page := resource.NewPage(triggers, resource.PageOptions(opts.PageOptions), nil)
	to := make([]*TFERunTrigger, len(page.Items))
	for i, t := range page.Items {
		to[i] = a.convert(t)
	}
	a.RespondWithPage(w, r, to, page.Pagination)
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

func (a *TFEAPI) includeWorkspace(ctx context.Context, v any) ([]any, error) {
	trigger, ok := v.(*TFERunTrigger)
	if !ok {
		return nil, nil
	}
	ws, err := a.Client.GetWorkspace(ctx, trigger.Workspace.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	include, err := workspace.ToTFE(a.Authorizer, ws, (&http.Request{}).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return []any{include}, nil
}

func (a *TFEAPI) includeSourceable(ctx context.Context, v any) ([]any, error) {
	trigger, ok := v.(*TFERunTrigger)
	if !ok {
		return nil, nil
	}
	ws, err := a.Client.GetWorkspace(ctx, trigger.Sourceable.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving sourceable workspace: %w", err)
	}
	include, err := workspace.ToTFE(a.Authorizer, ws, (&http.Request{}).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return []any{include}, nil
}
