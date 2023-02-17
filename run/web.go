package run

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/logs"
)

type web struct {
	logs.LogService
	otf.Renderer
	otf.WorkspaceService

	app app
}

type htmlLogChunk struct {
	otf.Chunk
}

func (h *web) addHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/runs", h.list)
	r.HandleFunc("/runs/{run_id}", h.get)
	r.HandleFunc("/runs/{run_id}/delete", h.delete)
	r.HandleFunc("/runs/{run_id}/cancel", h.cancel)
	r.HandleFunc("/runs/{run_id}/apply", h.apply)
	r.HandleFunc("/runs/{run_id}/discard", h.discard)

	// this handles the link the terraform CLI shows during a plan/apply.
	r.HandleFunc("/app/{organization_name}/{workspace_id}/runs/{run_id}", h.get)
}

func (h *web) list(w http.ResponseWriter, r *http.Request) {
	params := struct {
		ListOptions otf.ListOptions
		WorkspaceID string `schema:"workspace_id,required"`
	}{}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	runs, err := h.app.list(r.Context(), otf.RunListOptions{
		ListOptions: params.ListOptions,
		WorkspaceID: &params.WorkspaceID,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("run_list.tmpl", w, r, struct {
		*RunList
		otf.Workspace
		StreamID string
	}{
		RunList:   runs,
		Workspace: ws,
		StreamID:  "watch-ws-runs-" + otf.GenerateRandomString(5),
	})
}

func (h *web) get(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.app.get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.GetWorkspace(r.Context(), run.WorkspaceID())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get existing logs thus far received for each phase. If none are found then don't treat
	// that as an error because it merely means no logs have yet been received.
	planLogs, err := h.GetChunk(r.Context(), logs.GetChunkOptions{
		RunID: run.ID(),
		Phase: otf.PlanPhase,
	})
	if err != nil && !errors.Is(err, otf.ErrResourceNotFound) {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := h.GetChunk(r.Context(), logs.GetChunkOptions{
		RunID: run.ID(),
		Phase: otf.ApplyPhase,
	})
	if err != nil && !errors.Is(err, otf.ErrResourceNotFound) {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("run_get.tmpl", w, r, struct {
		*Run
		Workspace otf.Workspace
		PlanLogs  *htmlLogChunk
		ApplyLogs *htmlLogChunk
	}{
		Run:       run,
		Workspace: ws,
		PlanLogs:  &htmlLogChunk{planLogs},
		ApplyLogs: &htmlLogChunk{applyLogs},
	})
}

func (h *web) delete(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.app.get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.app.delete(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(run.WorkspaceID()), http.StatusFound)
}

func (h *web) cancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.app.get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.app.cancel(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Runs(run.WorkspaceID()), http.StatusFound)
}

func (h *web) apply(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.app.apply(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID)+"#apply", http.StatusFound)
}

func (h *web) discard(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.app.discard(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}
