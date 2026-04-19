package ui

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	runpkg "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

const LatestRunUpdate sseEvent = "LatestRunUpdate"

var _ = paths.Run // suppress unused import

type Handlers struct {
	authorizer authz.Interface
	logger     logr.Logger
	templates  *templates
	client     Client
}

type Client interface {
	CreateRun(context.Context, resource.TfeID, runpkg.CreateOptions) (*runpkg.Run, error)
	ListRuns(_ context.Context, opts runpkg.ListOptions) (*resource.Page[*runpkg.Run], error)
	GetRun(ctx context.Context, id resource.TfeID) (*runpkg.Run, error)
	GetChunk(ctx context.Context, opts runpkg.GetChunkOptions) (runpkg.Chunk, error)
	CancelRun(ctx context.Context, id resource.TfeID) error
	ForceCancelRun(ctx context.Context, id resource.TfeID) error
	DiscardRun(ctx context.Context, id resource.TfeID) error
	TailRun(context.Context, runpkg.TailOptions) (<-chan runpkg.Chunk, error)
	DeleteRun(context.Context, resource.TfeID) error
	ApplyRun(context.Context, resource.TfeID) error
	WatchRuns(ctx context.Context) (<-chan pubsub.Event[*runpkg.Event], func(), error)
	ListTriggeredRunIDs(ctx context.Context, runID resource.ID) ([]resource.TfeID, error)
	GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
	WatchWorkspaces(ctx context.Context) (<-chan pubsub.Event[*workspace.Event], func(), error)
	GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error)
	sourceIconGetter
}

type sourceIconGetter interface {
	GetSourceIcon(source source.Source) templ.Component
}

func NewHandlers(
	logger logr.Logger,
	client Client,
	authorizer authz.Interface,
) *Handlers {
	return &Handlers{
		logger:     logger,
		client:     client,
		authorizer: authorizer,
		templates: &templates{
			workspaces:  client,
			users:       client,
			sourceIcons: client,
		},
	}
}

func (h *Handlers) SingleRunTable(run *runpkg.Run) templ.Component {
	return h.templates.singleRunTable(run)
}

func (h *Handlers) AddHandlers(r *mux.Router) {
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	createdRun, err := h.client.CreateRun(r.Context(), params.WorkspaceID, runpkg.CreateOptions{
		IsDestroy: new(params.Operation == runpkg.DestroyAllOperation),
		PlanOnly:  new(params.Operation == runpkg.PlanOnlyOperation),
		Source:    source.UI,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	props := runListProps{
		status:              opts.Statuses,
		statusFilterVisible: opts.StatusFilterVisible,
		pageOptions:         opts.PageOptions,
	}

	var renderOptions []helpers.RenderPageOption
	if opts.ListOptions.WorkspaceID != nil {
		ws, err := h.client.GetWorkspace(r.Context(), *opts.WorkspaceID)
		if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
		renderOptions = append(renderOptions, helpers.WithWorkspace(ws, h.authorizer))
		props.filterByWorkspace = true
		props.canUpdateWorkspace = h.authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID)
	} else if opts.ListOptions.Organization != nil {
		renderOptions = append(
			renderOptions,
			helpers.WithOrganization(*opts.ListOptions.Organization),
		)
	} else {
		helpers.Error(r, w, "must provide either organization_name or workspace_id", helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	renderOptions = append(renderOptions, helpers.WithBreadcrumbs(
		helpers.Breadcrumb{Name: "Runs"},
	))

	page, err := h.client.ListRuns(r.Context(), opts.ListOptions)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	props.page = page

	helpers.RenderPage(
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	run, err := h.client.GetRun(r.Context(), runID)
	if err != nil {
		helpers.Error(r, w, "retrieving run: "+err.Error())
		return
	}

	ws, err := h.client.GetWorkspace(r.Context(), run.WorkspaceID)
	if err != nil {
		helpers.Error(r, w, "retrieving workspace: "+err.Error())
		return
	}

	// Get existing logs thus far received for each phase.
	planLogs, err := h.client.GetChunk(r.Context(), runpkg.GetChunkOptions{
		RunID: run.ID,
		Phase: runpkg.PlanPhase,
	})
	if err != nil {
		helpers.Error(r, w, "retrieving plan logs: "+err.Error())
		return
	}
	applyLogs, err := h.client.GetChunk(r.Context(), runpkg.GetChunkOptions{
		RunID: run.ID,
		Phase: runpkg.ApplyPhase,
	})
	if err != nil {
		helpers.Error(r, w, "retrieving apply logs: "+err.Error())
		return
	}

	// Get the IDs of any runs that are triggered as a result of this run. (they
	// are only triggered after a successful apply, so to avoid a db query check
	// that this run has been applied first).
	var triggeredRunIDs []resource.TfeID
	if run.Status == runstatus.Applied {
		triggeredRunIDs, err = h.client.ListTriggeredRunIDs(r.Context(), runID)
		if err != nil {
			helpers.Error(r, w, "retrieving run: "+err.Error())
			return
		}
	}

	props := getRunProps{
		run:             run,
		ws:              ws,
		planLogs:        runpkg.Chunk{Data: planLogs.Data},
		applyLogs:       runpkg.Chunk{Data: applyLogs.Data},
		triggeredRunIDs: triggeredRunIDs,
	}
	helpers.RenderPage(
		h.templates.getRun(props),
		run.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.authorizer),
		helpers.WithPreContent(getPreContent()),
		helpers.WithPostContent(getPostContent(props)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: props.run.ID.String()},
		),
	)
}

func (h *Handlers) deleteRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	runItem, err := h.client.GetRun(r.Context(), runID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	err = h.client.DeleteRun(r.Context(), runID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.Workspace(runItem.WorkspaceID), http.StatusFound)
}

func (h *Handlers) cancelRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.client.CancelRun(r.Context(), runID); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	w.Header().Add("HX-Redirect", paths.Run(runID))
}

func (h *Handlers) forceCancelRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.client.ForceCancelRun(r.Context(), runID); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	w.Header().Add("HX-Redirect", paths.Run(runID))
}

func (h *Handlers) applyRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.client.ApplyRun(r.Context(), runID); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(runID)+"#apply", http.StatusFound)
}

func (h *Handlers) discardRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if err := h.client.DiscardRun(r.Context(), runID); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	w.Header().Add("HX-Redirect", paths.Run(runID))
}

func (h *Handlers) retryRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	existingRun, err := h.client.GetRun(r.Context(), runID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	newRun, err := h.client.CreateRun(r.Context(), existingRun.WorkspaceID, runpkg.CreateOptions{
		ConfigurationVersionID: &existingRun.ConfigurationVersionID,
		IsDestroy:              &existingRun.IsDestroy,
		PlanOnly:               &existingRun.PlanOnly,
		Source:                 source.UI,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.Run(newRun.ID), http.StatusFound)
}

const (
	periodReportUpdate      sseEvent = "PeriodReportUpdate"
	runWidgetUpdate         sseEvent = "RunWidgetUpdate"
	runTimeUpdate           sseEvent = "RunTimeUpdate"
	planTimeUpdate          sseEvent = "PlanTimeUpdate"
	applyTimeUpdate         sseEvent = "ApplyTimeUpdate"
	planStatusUpdate        sseEvent = "PlanStatusUpdate"
	applyStatusUpdate       sseEvent = "ApplyStatusUpdate"
	triggeredRunAlertUpdate sseEvent = "TriggeredRunAlertUpdate"
)

func (h *Handlers) watchRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	conn := newSSEConnection(w, false)

	sub, _, err := h.client.WatchRuns(r.Context())
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	send := func() {
		run, err := h.client.GetRun(r.Context(), runID)
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
			if ev.Type == pubsub.CreatedEvent {
				if ev.Payload.TriggeringRunID != nil && *ev.Payload.TriggeringRunID == runID {
					if err := conn.Render(r.Context(), triggeredRunAlert(ev.Payload.ID), triggeredRunAlertUpdate); err != nil {
						return
					}
				}
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

func (h *Handlers) watchLatestRun(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// Setup event subscriptions first then retrieve workspace to ensure we
	// don't miss anything.
	workspacesSub, _, err := h.client.WatchWorkspaces(r.Context())
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	runsSub, _, err := h.client.WatchRuns(r.Context())
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	ws, err := h.client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	conn := newSSEConnection(w, false)

	// function for retrieving run, rendering fragment and sending to client.
	send := func(runID resource.TfeID) {
		run, err := h.client.GetRun(r.Context(), runID)
		if err != nil {
			// terminate conn on error
			return
		}
		if err := conn.Render(r.Context(), h.templates.singleRunTable(run), LatestRunUpdate); err != nil {
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
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ch, err := h.client.TailRun(r.Context(), params)
	if err != nil {
		helpers.Error(r, w, err.Error())
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
				h.logger.Error(err, "marshalling data")
				continue
			}
			conn.Send(js, EventLogChunk)
		case <-r.Context().Done():
			return
		}
	}
}
