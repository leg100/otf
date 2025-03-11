package run

import (
	"bytes"
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/components"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
)

type (
	webHandlers struct {
		logger               logr.Logger
		runs                 webRunClient
		workspaces           webWorkspaceClient
		authorizer           webAuthorizer
		websocketListHandler *components.WebsocketListHandler[*Run, ListOptions]
	}

	webRunClient interface {
		Create(ctx context.Context, workspaceID resource.ID, opts CreateOptions) (*Run, error)
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error)
		Get(ctx context.Context, id resource.ID) (*Run, error)
		Delete(ctx context.Context, runID resource.ID) error
		Cancel(ctx context.Context, runID resource.ID) error
		ForceCancel(ctx context.Context, runID resource.ID) error
		Apply(ctx context.Context, runID resource.ID) error
		Discard(ctx context.Context, runID resource.ID) error

		getLogs(ctx context.Context, runID resource.ID, phase internal.PhaseType) ([]byte, error)
		watchWithOptions(ctx context.Context, opts WatchOptions) (<-chan pubsub.Event[*Run], error)
	}

	webWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.ID) (*workspace.Workspace, error)
		GetWorkspacePolicy(ctx context.Context, workspaceID resource.ID) (authz.WorkspacePolicy, error)
	}

	webAuthorizer interface {
		CanAccess(context.Context, authz.Action, *authz.AccessRequest) bool
	}
)

func newWebHandlers(service *Service, opts Options) *webHandlers {
	return &webHandlers{
		authorizer: opts.Authorizer,
		logger:     opts.Logger,
		runs:       service,
		workspaces: opts.WorkspaceService,
		websocketListHandler: &components.WebsocketListHandler[*Run, ListOptions]{
			Logger: opts.Logger,
			Client: service,
			Tabular: &table{
				workspaceClient: opts.WorkspaceService,
			},
		},
	}
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/runs", h.list).Methods("GET")
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
		WorkspaceID resource.ID `schema:"workspace_id,required"`
		Operation   Operation   `schema:"operation,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := h.runs.Create(r.Context(), params.WorkspaceID, CreateOptions{
		IsDestroy: internal.Bool(params.Operation == DestroyAllOperation),
		PlanOnly:  internal.Bool(params.Operation == PlanOnlyOperation),
		Source:    SourceUI,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Workspace(params.WorkspaceID.String()), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID.String()), http.StatusFound)
}

func (h *webHandlers) list(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h.websocketListHandler.Handler(w, r)
		return
	}

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
	}

	if opts.ListOptions.WorkspaceID != nil {
		ws, err := h.workspaces.Get(r.Context(), *opts.WorkspaceID)
		if err != nil {
			html.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.organization = ws.Organization
		props.ws = ws
		props.canUpdateWorkspace = h.authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, &authz.AccessRequest{ID: &ws.ID})
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
	planLogs, err := h.runs.getLogs(r.Context(), run.ID, internal.PlanPhase)
	if err != nil {
		html.Error(w, "retrieving plan logs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := h.runs.getLogs(r.Context(), run.ID, internal.ApplyPhase)
	if err != nil {
		html.Error(w, "retrieving apply logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	props := getProps{
		run:       run,
		ws:        ws,
		planLogs:  logs.Chunk{Data: planLogs},
		applyLogs: logs.Chunk{Data: applyLogs},
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

	html.Render(widget(run), w, r)
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
	http.Redirect(w, r, paths.Workspace(run.WorkspaceID.String()), http.StatusFound)
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

	http.Redirect(w, r, paths.Run(runID.String()), http.StatusFound)
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

	http.Redirect(w, r, paths.Run(runID.String()), http.StatusFound)
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
	http.Redirect(w, r, paths.Run(runID.String())+"#apply", http.StatusFound)
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
	http.Redirect(w, r, paths.Run(runID.String()), http.StatusFound)
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
		Source:                 SourceUI,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Run(runID.String()), http.StatusFound)
		return
	}

	http.Redirect(w, r, paths.Run(run.ID.String()), http.StatusFound)
}

func (h *webHandlers) watch(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.ID  `schema:"workspace_id,required"`
		Latest      bool         `schema:"latest"`
		RunID       *resource.ID `schema:"run_id"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	events, err := h.runs.watchWithOptions(r.Context(), WatchOptions{
		WorkspaceID: &params.WorkspaceID,
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
			} else if params.RunID != nil && *params.RunID != event.Payload.ID {
				// skip: event is for a run which does not match the
				// filter
				continue
			}

			//
			// render HTML snippet and send as payload in SSE events
			//
			itemHTML := new(bytes.Buffer)
			if err := eventView(event.Payload).Render(r.Context(), itemHTML); err != nil {
				h.logger.Error(err, "rendering template for run item")
				continue
			}
			if event.Type == pubsub.CreatedEvent {
				// newly created run is sent with "created" event type
				pubsub.WriteSSEEvent(w, itemHTML.Bytes(), event.Type, false)
			} else {
				// updated run events target existing run items in page
				pubsub.WriteSSEEvent(w, itemHTML.Bytes(), pubsub.EventType("run-item-"+event.Payload.ID.String()), false)
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
