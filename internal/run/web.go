package run

import (
	"bytes"
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
)

type (
	webHandlers struct {
		html.Renderer

		logger     logr.Logger
		runs       webRunClient
		workspaces webWorkspaceClient
	}

	webRunClient interface {
		CreateRun(ctx context.Context, workspaceID string, opts CreateOptions) (*Run, error)
		ListRuns(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error)
		GetRun(ctx context.Context, id string) (*Run, error)
		Delete(ctx context.Context, runID string) error
		Cancel(ctx context.Context, runID string) error
		ForceCancelRun(ctx context.Context, runID string) error
		Apply(ctx context.Context, runID string) error
		DiscardRun(ctx context.Context, runID string) error
		Watch(ctx context.Context, opts WatchOptions) (<-chan pubsub.Event[*Run], error)
		getLogs(ctx context.Context, runID string, phase internal.PhaseType) ([]byte, error)
	}

	webWorkspaceClient interface {
		GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
		GetPolicy(ctx context.Context, workspaceID string) (internal.WorkspacePolicy, error)
	}
)

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/runs", h.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/start-run", h.createRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}", h.get).Methods("GET")
	r.HandleFunc("/runs/{run_id}/widget", h.getWidget).Methods("GET")
	r.HandleFunc("/runs/{run_id}/delete", h.delete).Methods("POST")
	r.HandleFunc("/runs/{run_id}/cancel", h.cancel).Methods("POST")
	r.HandleFunc("/runs/{run_id}/force-cancel", h.forceCancel).Methods("POST")
	r.HandleFunc("/runs/{run_id}/apply", h.apply).Methods("POST")
	r.HandleFunc("/runs/{run_id}/discard", h.discard).Methods("POST")
	r.HandleFunc("/runs/{run_id}/retry", h.retry).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/watch", h.watch).Methods("GET")

	// this handles the link the terraform CLI shows during a plan/apply.
	r.HandleFunc("/{organization_name}/{workspace_id}/runs/{run_id}", h.get).Methods("GET")
}

func (h *webHandlers) createRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string    `schema:"workspace_id,required"`
		Operation   Operation `schema:"operation,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.CreateRun(r.Context(), params.WorkspaceID, CreateOptions{
		IsDestroy: internal.Bool(params.Operation == DestroyAllOperation),
		PlanOnly:  internal.Bool(params.Operation == PlanOnlyOperation),
		Source:    SourceUI,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID), http.StatusFound)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		PageNumber  int    `schema:"page[number]"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.workspaces.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	policy, err := h.workspaces.GetPolicy(r.Context(), ws.ID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	runs, err := h.runs.ListRuns(r.Context(), ListOptions{
		WorkspaceID: &params.WorkspaceID,
		PageOptions: resource.PageOptions{
			PageNumber: params.PageNumber,
			PageSize:   html.PageSize,
		},
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		workspace.WorkspacePage
		*resource.Page[*Run]
		CanUpdateWorkspace bool
	}{
		WorkspacePage:      workspace.NewPage(r, "runs", ws),
		Page:               runs,
		CanUpdateWorkspace: user.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
	}

	if isHTMX := r.Header.Get("HX-Request"); isHTMX == "true" {
		if err := h.RenderTemplate("run_listing.tmpl", w, response); err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		h.Render("run_list.tmpl", w, response)
	}
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.GetRun(r.Context(), runID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.workspaces.GetWorkspace(r.Context(), run.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get existing logs thus far received for each phase.
	planLogs, err := h.runs.getLogs(r.Context(), run.ID, internal.PlanPhase)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := h.runs.getLogs(r.Context(), run.ID, internal.ApplyPhase)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("run_get.tmpl", w, struct {
		workspace.WorkspacePage
		Run       *Run
		PlanLogs  internal.Chunk
		ApplyLogs internal.Chunk
	}{
		WorkspacePage: workspace.NewPage(r, run.ID, ws),
		Run:           run,
		PlanLogs:      internal.Chunk{Data: planLogs},
		ApplyLogs:     internal.Chunk{Data: applyLogs},
	})
}

// getWidget renders a run "widget", i.e. the container that
// contains info about a run. Intended for use with an ajax request.
func (h *webHandlers) getWidget(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.GetRun(r.Context(), runID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.RenderTemplate("run_item.tmpl", w, run); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.GetRun(r.Context(), runID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.runs.Delete(r.Context(), runID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(run.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) cancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.runs.Cancel(r.Context(), runID); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *webHandlers) forceCancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.runs.ForceCancelRun(r.Context(), runID); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *webHandlers) apply(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.runs.Apply(r.Context(), runID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID)+"#apply", http.StatusFound)
}

func (h *webHandlers) discard(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.runs.DiscardRun(r.Context(), runID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *webHandlers) retry(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.GetRun(r.Context(), runID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	run, err = h.runs.CreateRun(r.Context(), run.WorkspaceID, CreateOptions{
		ConfigurationVersionID: &run.ConfigurationVersionID,
		IsDestroy:              &run.IsDestroy,
		PlanOnly:               &run.PlanOnly,
		Source:                 SourceUI,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Run(runID), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID), http.StatusFound)
}

func (h *webHandlers) watch(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		Latest      bool   `schema:"latest"`
		RunID       string `schema:"run_id"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	events, err := h.runs.Watch(r.Context(), WatchOptions{
		WorkspaceID: internal.String(params.WorkspaceID),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	rc := http.NewResponseController(w)
	rc.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			// Handle query parameters which filter run events:
			// - 'latest' specifies that the client is only interest in events
			// relating to the latest run for the workspace
			// - 'run-id' (mutually exclusive with 'latest') - specifies
			// that the client is only interested in events relating to that
			// run.
			// - otherwise, if neither of those parameters are specified
			// then events for all runs are relayed.
			if params.Latest && !event.Payload.Latest {
				// skip: run is not the latest run for a workspace
				continue
			} else if params.RunID != "" && params.RunID != event.Payload.ID {
				// skip: event is for a run which does not match the
				// filter
				continue
			}

			//
			// render HTML snippet and send as payload in SSE events
			//
			itemHTML := new(bytes.Buffer)
			if err := h.RenderTemplate("run_item.tmpl", itemHTML, event.Payload); err != nil {
				h.logger.Error(err, "rendering template for run item")
				continue
			}
			if event.Type == pubsub.CreatedEvent {
				// newly created run is sent with "created" event type
				pubsub.WriteSSEEvent(w, itemHTML.Bytes(), event.Type, false)
			} else {
				// updated run events target existing run items in page
				pubsub.WriteSSEEvent(w, itemHTML.Bytes(), pubsub.EventType("run-item-"+event.Payload.ID), false)
			}
			if params.Latest {
				// also write a 'latest-run' event if the caller has requested
				// the latest run for the workspace
				pubsub.WriteSSEEvent(w, itemHTML.Bytes(), "latest-run", false)
			}
			rc.Flush()
		}
	}
}
