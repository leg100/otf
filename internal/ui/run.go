package ui

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	runpkg "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
)

// addRunHandlers registers run UI handlers with the router
func addRunHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/organizations/{organization_name}/runs", h.listRuns).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/runs", h.listRuns).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/start-run", h.createRun).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/runs/watch-latest", h.watchLatestRun).Methods("GET")
	r.HandleFunc("/runs/{run_id}", h.getRun).Methods("GET")
	r.HandleFunc("/runs/{run_id}/delete", h.deleteRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}/cancel", h.cancelRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}/force-cancel", h.forceCancelRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}/apply", h.applyRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}/discard", h.discardRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}/retry", h.retryRun).Methods("POST")
	r.HandleFunc("/runs/{run_id}/watch", h.watchRun).Methods("GET")
	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)

	// this handles the link the terraform CLI shows during a plan/apply.
	r.HandleFunc("/{organization_name}/{workspace_id}/runs/{run_id}", h.getRun).Methods("GET")
}

func (h *Handlers) createRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID   `schema:"workspace_id,required"`
		Operation   runpkg.Operation `schema:"operation,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	createdRun, err := h.Runs.Create(r.Context(), params.WorkspaceID, runpkg.CreateOptions{
		IsDestroy: new(params.Operation == runpkg.DestroyAllOperation),
		PlanOnly:  new(params.Operation == runpkg.PlanOnlyOperation),
		Source:    source.UI,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(createdRun.ID), http.StatusFound)
}

func (h *Handlers) listRuns(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		runpkg.ListOptions
		StatusFilterVisible bool `schema:"status_filter_visible"`
	}
	if err := decode.All(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	props := runListProps{
		status:              opts.Statuses,
		statusFilterVisible: opts.StatusFilterVisible,
		pageOptions:         opts.PageOptions,
	}

	var renderOptions []renderPageOption
	if opts.ListOptions.WorkspaceID != nil {
		ws, err := h.Workspaces.Get(r.Context(), *opts.WorkspaceID)
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
		renderOptions = append(renderOptions, withWorkspace(ws))
		props.filterByWorkspace = true
		props.canUpdateWorkspace = h.Authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID)
	} else if opts.ListOptions.Organization != nil {
		renderOptions = append(
			renderOptions,
			withOrganization(*opts.ListOptions.Organization),
		)
	} else {
		html.Error(r, w, "must provide either organization_name or workspace_id", html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	renderOptions = append(renderOptions, withBreadcrumbs(
		helpers.Breadcrumb{Name: "Runs"},
	))

	page, err := h.Runs.List(r.Context(), opts.ListOptions)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	props.page = page

	h.renderPage(
		h.templates.runList(props),
		"runs",
		w,
		r,
		renderOptions...,
	)
}

func (h *Handlers) getRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	run, err := h.Runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(r, w, "retrieving run: "+err.Error())
		return
	}

	ws, err := h.Workspaces.Get(r.Context(), run.WorkspaceID)
	if err != nil {
		html.Error(r, w, "retrieving workspace: "+err.Error())
		return
	}

	// Get existing logs thus far received for each phase.
	planLogs, err := h.Runs.GetChunk(r.Context(), runpkg.GetChunkOptions{
		RunID: run.ID,
		Phase: runpkg.PlanPhase,
	})
	if err != nil {
		html.Error(r, w, "retrieving plan logs: "+err.Error())
		return
	}
	applyLogs, err := h.Runs.GetChunk(r.Context(), runpkg.GetChunkOptions{
		RunID: run.ID,
		Phase: runpkg.ApplyPhase,
	})
	if err != nil {
		html.Error(r, w, "retrieving apply logs: "+err.Error())
		return
	}

	props := getRunProps{
		run:       run,
		ws:        ws,
		planLogs:  runpkg.Chunk{Data: planLogs.Data},
		applyLogs: runpkg.Chunk{Data: applyLogs.Data},
	}
	h.renderPage(
		h.templates.getRun(props),
		run.ID.String(),
		w,
		r,
		withWorkspace(ws),
		withPreContent(getPreContent()),
		withPostContent(getPostContent(props)),
		withBreadcrumbs(
			helpers.Breadcrumb{Name: props.run.ID.String()},
		),
	)
}

func (h *Handlers) deleteRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	runItem, err := h.Runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	err = h.Runs.Delete(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.Workspace(runItem.WorkspaceID), http.StatusFound)
}

func (h *Handlers) cancelRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.Runs.Cancel(r.Context(), runID); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	w.Header().Add("HX-Redirect", paths.Run(runID))
}

func (h *Handlers) forceCancelRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.Runs.ForceCancel(r.Context(), runID); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	w.Header().Add("HX-Redirect", paths.Run(runID))
}

func (h *Handlers) applyRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.Runs.Apply(r.Context(), runID); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(runID)+"#apply", http.StatusFound)
}

func (h *Handlers) discardRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.Runs.Discard(r.Context(), runID); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	w.Header().Add("HX-Redirect", paths.Run(runID))
}

func (h *Handlers) retryRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	existingRun, err := h.Runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	newRun, err := h.Runs.Create(r.Context(), existingRun.WorkspaceID, runpkg.CreateOptions{
		ConfigurationVersionID: &existingRun.ConfigurationVersionID,
		IsDestroy:              &existingRun.IsDestroy,
		PlanOnly:               &existingRun.PlanOnly,
		Source:                 source.UI,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(newRun.ID), http.StatusFound)
}

const (
	periodReportUpdate sseEvent = "PeriodReportUpdate"
	runWidgetUpdate    sseEvent = "RunWidgetUpdate"
	runTimeUpdate      sseEvent = "RunTimeUpdate"
	planTimeUpdate     sseEvent = "PlanTimeUpdate"
	applyTimeUpdate    sseEvent = "ApplyTimeUpdate"
	planStatusUpdate   sseEvent = "PlanStatusUpdate"
	applyStatusUpdate  sseEvent = "ApplyStatusUpdate"
)

func (h *Handlers) watchRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	conn := newSSEConnection(w, false)

	sub, _ := h.Runs.Watch(r.Context())

	send := func() {
		run, err := h.Runs.Get(r.Context(), runID)
		if err != nil {
			// terminate conn on error
			return
		}
		// Render multiple html fragments each time a run event occurs. Each
		// fragment is sent down the SSE conn as separate SSE events.
		if err := conn.Render(r.Context(), runningTime(run), runTimeUpdate); err != nil {
			return
		}
		if err := conn.Render(r.Context(), runningTime(&run.Plan), planTimeUpdate); err != nil {
			return
		}
		if err := conn.Render(r.Context(), runningTime(&run.Apply), applyTimeUpdate); err != nil {
			return
		}
		if err := conn.Render(r.Context(), phaseStatus(run.Plan), planStatusUpdate); err != nil {
			return
		}
		if err := conn.Render(r.Context(), phaseStatus(run.Apply), applyStatusUpdate); err != nil {
			return
		}
		if err := conn.Render(r.Context(), periodReport(run), periodReportUpdate); err != nil {
			return
		}

		if err := conn.Render(r.Context(), h.templates.singleRunTable(run), runWidgetUpdate); err != nil {
			return
		}
	}
	// Immediately send fragments in case they've changed since the page was
	// first rendered.
	//
	// TODO: add versions to run resources and send rendered run version in
	// query param so that versions can be compared and this step can be
	// skipped.
	send()

	for {
		select {
		case ev, ok := <-sub:
			if !ok {
				return
			}
			if ev.Type == pubsub.DeletedEvent {
				// TODO: run has been deleted: user should be alerted and
				// client should not reconnect.
				return
			}
			if ev.Payload.ID != runID {
				continue
			}
			send()
		case <-r.Context().Done():
			return
		}
	}
}

const latestRunUpdate sseEvent = "LatestRunUpdate"

func (h *Handlers) watchLatestRun(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// Setup event subscriptions first then retrieve workspace to ensure we
	// don't miss anything.
	workspacesSub, _ := h.Workspaces.Watch(r.Context())
	runsSub, _ := h.Runs.Watch(r.Context())
	ws, err := h.Workspaces.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	conn := newSSEConnection(w, false)

	// function for retrieving run, rendering fragment and sending to client.
	send := func(runID resource.TfeID) {
		run, err := h.Runs.Get(r.Context(), runID)
		if err != nil {
			// terminate conn on error
			return
		}
		if err := conn.Render(r.Context(), h.templates.singleRunTable(run), latestRunUpdate); err != nil {
			// terminate conn on error
			return
		}
	}

	// maintain reference to ID of latest run for workspace.
	var latestRunID *resource.TfeID

	if ws.LatestRun != nil {
		latestRunID = &ws.LatestRun.ID
		// Immediately send fragment in case it's changed since the page was
		// first rendered.
		//
		// TODO: add versions to run resources and send rendered run version in
		// query param so that versions can be compared and this step can be
		// skipped.
		send(*latestRunID)
	}

	for {
		select {
		case event, ok := <-workspacesSub:
			if !ok {
				return
			}
			if event.Payload.ID != workspaceID {
				// Event is for a different workspace, so skip.
				continue
			}
			if event.Payload.LatestRunID == nil {
				// Workspace doesn't have a latest run, so nothing to send to
				// client
				continue
			}
			if event.Payload.LatestRunID == latestRunID {
				// Workspace's latest run hasn't changed so nothing new to send
				// to client.
				continue
			}
			latestRunID = event.Payload.LatestRunID
		case event, ok := <-runsSub:
			if !ok {
				return
			}
			if latestRunID == nil {
				// Workspace doesn't have a latest run, so nothing to send to
				// client
				continue
			}
			if event.Payload.ID != *latestRunID {
				// Event is for a run different than the workspace's latest run,
				// so ignore.
				continue
			}
		case <-r.Context().Done():
			return
		}
		send(*latestRunID)
	}
}

const (
	EventLogChunk    sseEvent = "log_update"
	EventLogFinished sseEvent = "log_finished"
)

func (h *Handlers) tailRun(w http.ResponseWriter, r *http.Request) {
	var params runpkg.TailOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ch, err := h.Runs.Tail(r.Context(), params)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	conn := newSSEConnection(w, false)

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				// no more logs
				conn.Send([]byte("no more logs"), EventLogFinished)
				return
			}
			html := chunk.ToHTML()
			if len(html) == 0 {
				// don't send empty chunks
				continue
			}
			js, err := json.Marshal(struct {
				HTML       string `json:"html"`
				NextOffset int    `json:"offset"`
			}{
				HTML:       string(html) + "<br>",
				NextOffset: chunk.NextOffset(),
			})
			if err != nil {
				h.Logger.Error(err, "marshalling data")
				continue
			}
			conn.Send(js, EventLogChunk)
		case <-r.Context().Done():
			return
		}
	}
}
