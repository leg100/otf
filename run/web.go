package run

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/workspace"
)

type (
	webHandlers struct {
		otf.LogService
		otf.Renderer
		workspace.WorkspaceService

		starter runStarter
		svc     service
	}

	runStarter interface {
		startRun(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*Run, error)
	}
)

func (h *webHandlers) addHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/runs", h.list)
	r.HandleFunc("/workspaces/{workspace_id}/start-run", h.startRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}", h.get)
	r.HandleFunc("/runs/{run_id}/delete", h.delete)
	r.HandleFunc("/runs/{run_id}/cancel", h.cancel)
	r.HandleFunc("/runs/{run_id}/apply", h.apply)
	r.HandleFunc("/runs/{run_id}/discard", h.discard)

	// this handles the link the terraform CLI shows during a plan/apply.
	r.HandleFunc("/app/{organization_name}/{workspace_id}/runs/{run_id}", h.get)
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
	runs, err := h.svc.list(r.Context(), RunListOptions{
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

	var opts otf.ConfigurationVersionCreateOptions
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
