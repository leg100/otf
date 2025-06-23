package run

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
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

type (
	webHandlers struct {
		logger     logr.Logger
		runs       webRunClient
		workspaces webWorkspaceClient
		users      webUsersClient
		authorizer webAuthorizer
	}

	webRunClient interface {
		Create(ctx context.Context, workspaceID resource.TfeID, opts CreateOptions) (*Run, error)
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error)
		Get(ctx context.Context, id resource.TfeID) (*Run, error)
		Delete(ctx context.Context, runID resource.TfeID) error
		Cancel(ctx context.Context, runID resource.TfeID) error
		ForceCancel(ctx context.Context, runID resource.TfeID) error
		Apply(ctx context.Context, runID resource.TfeID) error
		Discard(ctx context.Context, runID resource.TfeID) error
		Watch(ctx context.Context) (<-chan pubsub.Event[*Event], func())
		Tail(ctx context.Context, opts TailOptions) (<-chan Chunk, error)
		GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	}

	webWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
		Watch(ctx context.Context) (<-chan pubsub.Event[*workspace.Event], func())
	}

	webUsersClient interface {
		GetUser(ctx context.Context, username user.UserSpec) (*user.User, error)
	}

	webAuthorizer interface {
		CanAccess(context.Context, authz.Action, resource.ID) bool
	}
)

func newWebHandlers(service *Service, opts Options) *webHandlers {
	return &webHandlers{
		authorizer: opts.Authorizer,
		logger:     opts.Logger,
		runs:       service,
		workspaces: opts.WorkspaceService,
		users:      opts.UsersService,
	}
}

func (h *webHandlers) addHandlers(r *mux.Router) {
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

func (h *webHandlers) createRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		Operation   Operation      `schema:"operation,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.Create(r.Context(), params.WorkspaceID, CreateOptions{
		IsDestroy: internal.Ptr(params.Operation == DestroyAllOperation),
		PlanOnly:  internal.Ptr(params.Operation == PlanOnlyOperation),
		Source:    source.UI,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID), http.StatusFound)
}

func (h *webHandlers) listByOrganization(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h := &components.WebsocketListHandler[*Run, *Event, ListOptions]{
			Logger: h.logger,
			Client: h.runs,
			Populator: table{
				workspaceClient: h.workspaces,
				users:           h.users,
			},
			ID: "page-results",
		}
		h.Handler(w, r)
		return
	}
	h.list(w, r)
}

func (h *webHandlers) listByWorkspace(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h := &components.WebsocketListHandler[*Run, *Event, ListOptions]{
			Logger: h.logger,
			Client: h.runs,
			Populator: table{
				users: h.users,
			},
			ID: "page-results",
		}
		h.Handler(w, r)
		return
	}
	h.list(w, r)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		ListOptions
		StatusFilterVisible bool `schema:"status_filter_visible"`
	}
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	props := listProps{
		status:              opts.Statuses,
		statusFilterVisible: opts.StatusFilterVisible,
		pageOptions:         opts.PageOptions,
	}

	if opts.ListOptions.WorkspaceID != nil {
		ws, err := h.workspaces.Get(r.Context(), *opts.WorkspaceID)
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.organization = ws.Organization
		props.ws = ws
		props.canUpdateWorkspace = h.authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID)
	} else if opts.ListOptions.Organization != nil {
		props.organization = *opts.ListOptions.Organization
	} else {
		html.Error(w, "must provide either organization_name or workspace_id", http.StatusUnprocessableEntity)
		return
	}

	html.Render(list(props), w, r)
}

func (h *webHandlers) get(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(w, "retrieving run: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ws, err := h.workspaces.Get(r.Context(), run.WorkspaceID)
	if err != nil {
		html.Error(w, "retrieving workspace: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get existing logs thus far received for each phase.
	planLogs, err := h.runs.GetChunk(r.Context(), GetChunkOptions{
		RunID: run.ID,
		Phase: PlanPhase,
	})
	if err != nil {
		html.Error(w, "retrieving plan logs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := h.runs.GetChunk(r.Context(), GetChunkOptions{
		RunID: run.ID,
		Phase: ApplyPhase,
	})
	if err != nil {
		html.Error(w, "retrieving apply logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	props := getProps{
		run:       run,
		ws:        ws,
		planLogs:  Chunk{Data: planLogs.Data},
		applyLogs: Chunk{Data: applyLogs.Data},
	}
	html.Render(get(props), w, r)
}

// getWidget renders a run "widget", i.e. the container that
// contains info about a run. Intended for use with an ajax request.
func (h *webHandlers) getWidget(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	table := components.UnpaginatedTable(
		&table{users: h.users},
		[]*Run{run},
		"run-item-"+run.ID.String(),
	)

	html.Render(table, w, r)
}

func (h *webHandlers) delete(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.runs.Delete(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(run.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) cancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.runs.Cancel(r.Context(), runID); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *webHandlers) forceCancel(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.runs.ForceCancel(r.Context(), runID); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *webHandlers) apply(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.runs.Apply(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID)+"#apply", http.StatusFound)
}

func (h *webHandlers) discard(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.runs.Discard(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(runID), http.StatusFound)
}

func (h *webHandlers) retry(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.Get(r.Context(), runID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	run, err = h.runs.Create(r.Context(), run.WorkspaceID, CreateOptions{
		ConfigurationVersionID: &run.ConfigurationVersionID,
		IsDestroy:              &run.IsDestroy,
		PlanOnly:               &run.PlanOnly,
		Source:                 source.UI,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Run(runID), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID), http.StatusFound)
}

func (h *webHandlers) watchRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !websocket.IsWebSocketUpgrade(r) {
		return
	}
	// Render a one-row table containing run each time a run event arrives.
	comp := func(run *Run) templ.Component {
		return components.UnpaginatedTable(
			&table{users: h.users},
			[]*Run{run},
			"run-item-"+run.ID.String(),
		)
	}
	conn, err := components.NewWebsocket(h.logger, w, r, h.runs, comp)
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

func (h *webHandlers) watchLatest(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !websocket.IsWebSocketUpgrade(r) {
		return
	}
	conn, err := components.NewWebsocket(
		h.logger, w, r,
		h.runs,
		func(run *Run) templ.Component {
			return components.UnpaginatedTable(
				&table{users: h.users},
				[]*Run{run},
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
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
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

func (h *webHandlers) tailRun(w http.ResponseWriter, r *http.Request) {
	var params TailOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ch, err := h.runs.Tail(r.Context(), params)
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
