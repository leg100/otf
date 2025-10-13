package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/components"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

type (
	runHandlers struct {
		logger     logr.Logger
		runs       runClient
		workspaces runWorkspaceClient
		users      runUsersClient
		authorizer runAuthorizer
	}

	runClient interface {
		Create(ctx context.Context, workspaceID resource.TfeID, opts run.CreateOptions) (*run.Run, error)
		List(ctx context.Context, opts run.ListOptions) (*resource.Page[*run.Run], error)
		Get(ctx context.Context, id resource.TfeID) (*run.Run, error)
		Delete(ctx context.Context, runID resource.TfeID) error
		Cancel(ctx context.Context, runID resource.TfeID) error
		ForceCancel(ctx context.Context, runID resource.TfeID) error
		Apply(ctx context.Context, runID resource.TfeID) error
		Discard(ctx context.Context, runID resource.TfeID) error
		Watch(ctx context.Context) (<-chan pubsub.Event[*run.Event], func())
		Tail(ctx context.Context, opts run.TailOptions) (<-chan run.Chunk, error)
		GetChunk(ctx context.Context, opts run.GetChunkOptions) (run.Chunk, error)
	}

	runWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
		Watch(ctx context.Context) (<-chan pubsub.Event[*workspace.Event], func())
	}

	runWorkspaceGetClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	}

	runUsersClient interface {
		GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error)
	}

	runAuthorizer interface {
		CanAccess(context.Context, authz.Action, resource.ID) bool
	}
)

// AddRunHandlers registers run UI handlers with the router
func AddRunHandlers(r *mux.Router, logger logr.Logger, runs runClient, workspaces runWorkspaceClient, users runUsersClient, authorizer runAuthorizer) {
	h := &runHandlers{
		authorizer: authorizer,
		logger:     logger,
		runs:       runs,
		workspaces: workspaces,
		users:      users,
	}
	h.addHandlers(r)
}

func (h *runHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/runs", h.listByOrganization).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/runs", h.listByWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/start-run", h.createRun).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/runs/watch-latest", h.watchLatest).Methods("GET")
	r.HandleFunc("/runs/{run_id}", h.get).Methods("GET")
	r.HandleFunc("/runs/{run_id}/widget", h.getWidget).Methods("GET")
	r.HandleFunc("/runs/{run_id}/delete", h.delete).Methods("POST")
	r.HandleFunc("/runs/{run_id}/cancel", h.cancel).Methods("POST")
	r.HandleFunc("/runs/{run_id}/force-cancel", h.forceCancel).Methods("POST")
	r.HandleFunc("/runs/{run_id}/apply", h.apply).Methods("POST")
	r.HandleFunc("/runs/{run_id}/discard", h.discard).Methods("POST")
	r.HandleFunc("/runs/{run_id}/retry", h.retry).Methods("POST")
	r.HandleFunc("/runs/{run_id}/watch", h.watchRun).Methods("GET")
	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)

	// this handles the link the terraform CLI shows during a plan/apply.
	r.HandleFunc("/{organization_name}/{workspace_id}/runs/{run_id}", h.get).Methods("GET")
}

func (h *runHandlers) createRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID   `schema:"workspace_id,required"`
		Operation   run.Operation    `schema:"operation,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	createdRun, err := h.runs.Create(r.Context(), params.WorkspaceID, run.CreateOptions{
		IsDestroy: internal.Ptr(params.Operation == run.DestroyAllOperation),
		PlanOnly:  internal.Ptr(params.Operation == run.PlanOnlyOperation),
		Source:    source.UI,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(createdRun.ID), http.StatusFound)
}

func (h *runHandlers) listByOrganization(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h := &components.WebsocketListHandler[*run.Run, *run.Event, run.ListOptions]{
			Logger: h.logger,
			Client: h.runs,
			Populator: runTable{
				workspaceGetClient: newWorkspaceCache(h.workspaces),
				users:           newUserCache(h.users),
			},
			ID: "page-results",
		}
		h.Handler(w, r)
		return
	}
	h.list(w, r)
}

func (h *runHandlers) listByWorkspace(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h := &components.WebsocketListHandler[*run.Run, *run.Event, run.ListOptions]{
			Logger: h.logger,
			Client: h.runs,
			Populator: runTable{
				users: newUserCache(h.users),
			},
			ID: "page-results",
		}
		h.Handler(w, r)
		return
	}
	h.list(w, r)
}

func (h *runHandlers) list(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		run.ListOptions
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

	if opts.ListOptions.WorkspaceID != nil {
		ws, err := h.workspaces.Get(r.Context(), *opts.WorkspaceID)
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
		props.organization = ws.Organization
		props.ws = ws
		props.canUpdateWorkspace = h.authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID)
	} else if opts.ListOptions.Organization != nil {
		props.organization = *opts.ListOptions.Organization
	} else {
		html.Error(r, w, "must provide either organization_name or workspace_id", html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	html.Render(runList(props), w, r)
}

func (h *runHandlers) get(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	runResult, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(r, w, "retrieving run: "+err.Error())
		return
	}

	ws, err := h.workspaces.Get(r.Context(), runResult.WorkspaceID)
	if err != nil {
		html.Error(r, w, "retrieving workspace: "+err.Error())
		return
	}

	// Get existing logs thus far received for each phase.
	planLogs, err := h.runs.GetChunk(r.Context(), run.GetChunkOptions{
		RunID: runResult.ID,
		Phase: run.PlanPhase,
	})
	if err != nil {
		html.Error(r, w, "retrieving plan logs: "+err.Error())
		return
	}
	applyLogs, err := h.runs.GetChunk(r.Context(), run.GetChunkOptions{
		RunID: runResult.ID,
		Phase: run.ApplyPhase,
	})
	if err != nil {
		html.Error(r, w, "retrieving apply logs: "+err.Error())
		return
	}

	props := runGetProps{
		run:       runResult,
		ws:        ws,
		planLogs:  run.Chunk{Data: planLogs.Data},
		applyLogs: run.Chunk{Data: applyLogs.Data},
	}
	html.Render(runGet(props), w, r)
}

// getWidget renders a run "widget", i.e. the container that
// contains info about a run. Intended for use with an ajax request.
func (h *runHandlers) getWidget(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	runItem, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	table := components.UnpaginatedTable(
		runTable{users: h.users},
		[]*run.Run{runItem},
		"run-item-"+runItem.ID.String(),
	)

	html.Render(table, w, r)
}

func (h *runHandlers) delete(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	runItem, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	err = h.runs.Delete(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.Workspace(runItem.WorkspaceID), http.StatusFound)
}

func (h *runHandlers) cancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.runs.Cancel(r.Context(), runID); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *runHandlers) forceCancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.runs.ForceCancel(r.Context(), runID); err != nil {
		html.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *runHandlers) apply(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err = h.runs.Apply(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.Run(runID)+"#apply", http.StatusFound)
}

func (h *runHandlers) discard(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err = h.runs.Discard(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *runHandlers) retry(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	existingRun, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	newRun, err := h.runs.Create(r.Context(), existingRun.WorkspaceID, run.CreateOptions{
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

func (h *runHandlers) watchRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	if !websocket.IsWebSocketUpgrade(r) {
		return
	}
	// Render a one-row table containing run each time a run event arrives.
	conn, err := components.NewWebsocket(h.logger, w, r, h.runs, (&event{users: h.users}).view)
	if err != nil {
		h.logger.Error(err, "upgrading websocket connection")
		return
	}
	defer conn.Close()

	sub, _ := h.runs.Watch(r.Context())

	if !conn.Send(runID) {
		return
	}

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case event := <-sub:
			if event.Type == pubsub.DeletedEvent {
				// TODO: run has been deleted: user should be alerted and
				// client should not reconnect.
				return
			}
			if event.Payload.ID != runID {
				continue
			}
		case <-r.Context().Done():
			return
		}
		// all further run events currently waiting on the subscription
		// channel are rendered redundant because the websocket client
		// retrieves the latest version of the run before sending it.
		for {
			select {
			case <-sub:
			default:
				goto done
			}
		}
	done:
		if !conn.Send(runID) {
			return
		}
		// Wait before sending anything more to client to avoid sending too many
		// messages.
		select {
		case <-ticker.C:
		case <-r.Context().Done():
			return
		}
	}
}

func (h *runHandlers) watchLatest(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	if !websocket.IsWebSocketUpgrade(r) {
		return
	}
	conn, err := components.NewWebsocket(
		h.logger, w, r,
		h.runs,
		func(runItem *run.Run) templ.Component {
			return components.UnpaginatedTable(
				runTable{users: h.users},
				[]*run.Run{runItem},
				"latest-run",
			)
		},
	)
	if err != nil {
		h.logger.Error(err, "upgrading websocket connection")
		return
	}
	defer conn.Close()
	// Setup event subscriptions first then retrieve workspace to ensure we
	// don't miss anything.
	workspacesSub, _ := h.workspaces.Watch(r.Context())
	runsSub, _ := h.runs.Watch(r.Context())
	ws, err := h.workspaces.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	var latestRunID *resource.TfeID
	if ws.LatestRun != nil {
		latestRunID = &ws.LatestRun.ID
		if !conn.Send(*latestRunID) {
			return
		}
	}
	for {
		select {
		case event := <-workspacesSub:
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
		case event := <-runsSub:
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
		if !conn.Send(*latestRunID) {
			return
		}
	}
}

const (
	EventLogChunk    string = "log_update"
	EventLogFinished string = "log_finished"
)

func (h *runHandlers) tailRun(w http.ResponseWriter, r *http.Request) {
	var params run.TailOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ch, err := h.runs.Tail(r.Context(), params)
	if err != nil {
		html.Error(r, w, err.Error())
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
		case chunk, ok := <-ch:
			if !ok {
				// no more logs
				pubsub.WriteSSEEvent(w, []byte("no more logs"), EventLogFinished, false)
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
				h.logger.Error(err, "marshalling data")
				continue
			}
			pubsub.WriteSSEEvent(w, js, EventLogChunk, false)
			rc.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
