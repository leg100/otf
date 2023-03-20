package run

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/workspace"
)

type (
	webHandlers struct {
		logr.Logger
		otf.LogService
		otf.Renderer
		WorkspaceService

		starter runStarter
		svc     Service
	}

	runStarter interface {
		startRun(ctx context.Context, workspaceID string, opts configversion.ConfigurationVersionCreateOptions) (*Run, error)
	}
)

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/runs", h.list)
	r.HandleFunc("/workspaces/{workspace_id}/start-run", h.startRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}", h.get)
	r.HandleFunc("/runs/{run_id}/delete", h.delete)
	r.HandleFunc("/runs/{run_id}/cancel", h.cancel)
	r.HandleFunc("/runs/{run_id}/apply", h.apply)
	r.HandleFunc("/runs/{run_id}/discard", h.discard)
	r.HandleFunc("/workspaces/{workspace_id}/watch", h.watchWorkspace).Methods("GET")

	// this handles the link the terraform CLI shows during a plan/apply.
	r.HandleFunc("/{organization_name}/{workspace_id}/runs/{run_id}", h.get)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	var params struct {
		otf.ListOptions
		WorkspaceID string `schema:"workspace_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	runs, err := h.svc.ListRuns(r.Context(), RunListOptions{
		ListOptions: params.ListOptions,
		WorkspaceID: &params.WorkspaceID,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("run_list.tmpl", w, r, struct {
		*RunList
		*workspace.Workspace
		StreamID string
	}{
		RunList:   runs,
		Workspace: ws,
		StreamID:  "watch-ws-runs-" + otf.GenerateRandomString(5),
	})
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.svc.get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.GetWorkspace(r.Context(), run.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get existing logs thus far received for each phase. If none are found then don't treat
	// that as an error because it merely means no logs have yet been received.
	planLogs, err := h.GetChunk(r.Context(), otf.GetChunkOptions{
		RunID: run.ID,
		Phase: otf.PlanPhase,
	})
	if err != nil && !errors.Is(err, otf.ErrResourceNotFound) {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := h.GetChunk(r.Context(), otf.GetChunkOptions{
		RunID: run.ID,
		Phase: otf.ApplyPhase,
	})
	if err != nil && !errors.Is(err, otf.ErrResourceNotFound) {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("run_get.tmpl", w, r, struct {
		*Run
		Workspace *workspace.Workspace
		PlanLogs  otf.Chunk
		ApplyLogs otf.Chunk
	}{
		Run:       run,
		Workspace: ws,
		PlanLogs:  planLogs,
		ApplyLogs: applyLogs,
	})
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.svc.get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.svc.delete(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(run.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) cancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.svc.get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.svc.cancel(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Runs(run.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) apply(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.svc.apply(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID)+"#apply", http.StatusFound)
}

func (h *webHandlers) discard(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.svc.discard(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *webHandlers) startRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		Strategy    string `schema:"strategy,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var opts configversion.ConfigurationVersionCreateOptions
	switch params.Strategy {
	case "plan-only":
		opts.Speculative = otf.Bool(true)
	case "plan-and-apply":
		opts.Speculative = otf.Bool(false)
	default:
		html.Error(w, "invalid strategy", http.StatusUnprocessableEntity)
		return
	}

	run, err := h.starter.startRun(r.Context(), params.WorkspaceID, opts)
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID), http.StatusFound)
}

func (h *webHandlers) watchWorkspace(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID string `schema:"workspace_id,required"`
		StreamID    string `schema:"stream,required"`
		Latest      bool   `schema:"latest"`
		RunID       string `schema:"run_id"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	events, err := h.svc.Watch(r.Context(), WatchOptions{
		WorkspaceID: otf.String(params.WorkspaceID),
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
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
			run, ok := event.Payload.(*Run)
			if !ok {
				// skip non-run events
				continue
			}

			// Handle query parameters which filter run events:
			// - 'latest' specifies that the client is only interest in events
			// relating to the latest run for the workspace
			// - 'run-id' (mutually exclusive with 'latest') - specifies
			// that the client is only interested in events relating to that
			// run.
			// - otherwise, if neither of those parameters are specified
			// then events for all runs are relayed.
			if params.Latest && !run.Latest {
				// skip: run is not the latest run for a workspace
				continue
			} else if params.RunID != "" && params.RunID != run.ID {
				// skip: event is for a run which does not match the
				// filter
				continue
			}

			// render HTML snippets and send as payload in SSE event
			itemHTML := new(bytes.Buffer)
			if err := h.RenderTemplate("run_item.tmpl", itemHTML, run); err != nil {
				h.Error(err, "rendering template for run item")
				continue
			}
			runStatusHTML := new(bytes.Buffer)
			if err := h.RenderTemplate("run_status.tmpl", runStatusHTML, run); err != nil {
				h.Error(err, "rendering run status template")
				continue
			}
			planStatusHTML := new(bytes.Buffer)
			if err := h.RenderTemplate("phase_status.tmpl", planStatusHTML, run.Plan); err != nil {
				h.Error(err, "rendering plan status template")
				continue
			}
			applyStatusHTML := new(bytes.Buffer)
			if err := h.RenderTemplate("phase_status.tmpl", applyStatusHTML, run.Apply); err != nil {
				h.Error(err, "rendering apply status template")
				continue
			}
			js, err := json.Marshal(struct {
				ID              string        `json:"id"`
				RunStatus       otf.RunStatus `json:"run-status"`
				RunItemHTML     string        `json:"run-item-html"`
				RunStatusHTML   string        `json:"run-status-html"`
				PlanStatusHTML  string        `json:"plan-status-html"`
				ApplyStatusHTML string        `json:"apply-status-html"`
			}{
				ID:              run.ID,
				RunStatus:       run.Status,
				RunItemHTML:     itemHTML.String(),
				RunStatusHTML:   runStatusHTML.String(),
				PlanStatusHTML:  planStatusHTML.String(),
				ApplyStatusHTML: applyStatusHTML.String(),
			})
			if err != nil {
				h.Error(err, "marshalling watched run", "run", run.ID)
				continue
			}
			otf.WriteSSEEvent(w, js, event.Type)
			rc.Flush()
		}
	}
}
