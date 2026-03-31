package run

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

var tfeUser = resource.MustHardcodeTfeID(resource.UserKind, "123")

type tfe struct {
	*tfeapi.Responder
	client     tfeClient
	signer     tfeapi.Signer
	authorizer *authz.Authorizer
}

type tfeClient interface {
	CreateRun(context.Context, resource.TfeID, CreateOptions) (*Run, error)
	ListRuns(_ context.Context, opts ListOptions) (*resource.Page[*Run], error)
	GetRun(ctx context.Context, id resource.TfeID) (*Run, error)
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	CancelRun(ctx context.Context, id resource.TfeID) error
	ForceCancelRun(ctx context.Context, id resource.TfeID) error
	DiscardRun(ctx context.Context, id resource.TfeID) error
	TailRun(context.Context, TailOptions) (<-chan Chunk, error)
	DeleteRun(context.Context, resource.TfeID) error
	ApplyRun(context.Context, resource.TfeID) error
	WatchRuns(ctx context.Context) (<-chan pubsub.Event[*Event], func())
	GetRunPlanFile(ctx context.Context, id resource.TfeID, format PlanFormat) ([]byte, error)
	GetWorkspace(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
}

func NewTFEAPI(
	client tfeClient,
	authorizer *authz.Authorizer,
	signer tfeapi.Signer,
	responder *tfeapi.Responder,
) *tfe {
	api := &tfe{
		client:     client,
		authorizer: authorizer,
		signer:     signer,
		Responder:  responder,
	}
	// Fetch related resources when API requests their inclusion
	responder.Register(tfeapi.IncludeCreatedBy, api.includeCreatedBy)
	responder.Register(tfeapi.IncludeCurrentRun, api.includeCurrentRun)
	responder.Register(tfeapi.IncludeWorkspace, api.includeWorkspace)
	return api
}

func (a *tfe) AddHandlers(r *mux.Router) {
	// Run routes
	r.HandleFunc("/runs", a.createRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/apply", a.applyRun).Methods("POST")
	r.HandleFunc("/runs", a.listRuns).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/runs", a.listRuns).Methods("GET")
	r.HandleFunc("/runs/{id}", a.getRun).Methods("GET")
	r.HandleFunc("/runs/{id}/actions/discard", a.discardRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/cancel", a.cancelRun).Methods("POST")
	r.HandleFunc("/runs/{id}/actions/force-cancel", a.forceCancelRun).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/runs/queue", a.getRunQueue).Methods("GET")

	// Plan routes
	r.HandleFunc("/plans/{plan_id}", a.getPlan).Methods("GET")
	r.HandleFunc("/plans/{plan_id}/json-output", a.getPlanJSON).Methods("GET")

	// Apply routes
	r.HandleFunc("/applies/{apply_id}", a.getApply).Methods("GET")

	// Run events routes
	r.HandleFunc("/runs/{id}/run-events", a.listRunEvents).Methods("GET")

	// Log routes
	//
	// getLogs is not part of the documented TFEAPI, but called by the
	// terraform cli with a signed URL
	r.HandleFunc("/runs/{run_id}/logs/{phase}", a.getLogs).Methods("GET")
}

func (a *tfe) createRun(w http.ResponseWriter, r *http.Request) {
	var params TFERunCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if params.Workspace == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "workspace"})
		return
	}

	opts := CreateOptions{
		AutoApply:        params.AutoApply,
		IsDestroy:        params.IsDestroy,
		Refresh:          params.Refresh,
		RefreshOnly:      params.RefreshOnly,
		Message:          params.Message,
		TargetAddrs:      params.TargetAddrs,
		ReplaceAddrs:     params.ReplaceAddrs,
		PlanOnly:         params.PlanOnly,
		Source:           source.API,
		AllowEmptyApply:  params.AllowEmptyApply,
		TerraformVersion: params.TerraformVersion,
	}
	if params.ConfigurationVersion != nil {
		opts.ConfigurationVersionID = &params.ConfigurationVersion.ID
	}
	if tfeapi.IsTerraformCLI(r) {
		opts.Source = source.Terraform
	}
	opts.Variables = make([]Variable, len(params.Variables))
	for i, from := range params.Variables {
		opts.Variables[i] = Variable{Key: from.Key, Value: from.Value}
	}

	run, err := a.client.CreateRun(r.Context(), params.Workspace.ID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	converted, err := a.toRun(run, r.Context())
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, converted, http.StatusCreated)
}

func (a *tfe) getRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	run, err := a.client.GetRun(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	converted, err := a.toRun(run, r.Context())
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, converted, http.StatusOK)
}

func (a *tfe) listRuns(w http.ResponseWriter, r *http.Request) {
	var params TFERunListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert comma-separated list of statuses to []RunStatus
	statuses := internal.FromStringCSV[runstatus.Status](params.Status)
	// convert comma-separated list of sources to []RunSource
	sources := internal.FromStringCSV[source.Source](params.Source)
	// split operations CSV
	operations := internal.SplitCSV(params.Operation)
	var planOnly *bool
	if slices.Contains(operations, string(RunOperationPlanOnly)) {
		planOnly = new(true)
	}

	a.listRunsWithOptions(w, r, ListOptions{
		Organization: params.Organization,
		WorkspaceID:  params.WorkspaceID,
		PageOptions:  resource.PageOptions(params.ListOptions),
		Statuses:     statuses,
		Sources:      sources,
		PlanOnly:     planOnly,
		CommitSHA:    params.Commit,
		VCSUsername:  params.User,
	})
}

func (a *tfe) getRunQueue(w http.ResponseWriter, r *http.Request) {
	a.listRunsWithOptions(w, r, ListOptions{
		Statuses: []runstatus.Status{runstatus.PlanQueued, runstatus.ApplyQueued},
	})
}

func (a *tfe) listRunsWithOptions(w http.ResponseWriter, r *http.Request, opts ListOptions) {
	page, err := a.client.ListRuns(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*TFERun, len(page.Items))
	for i, from := range page.Items {
		to, err := a.toRun(from, r.Context())
		if err != nil {
			tfeapi.Error(w, err)
			return
		}
		items[i] = to
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) applyRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.client.ApplyRun(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *tfe) discardRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err = a.client.DiscardRun(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *tfe) cancelRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err = a.client.CancelRun(r.Context(), id); err != nil {
		if internal.ErrorIs(err, ErrRunCancelNotAllowed, ErrRunForceCancelNotAllowed) {
			tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusConflict))
		} else {
			tfeapi.Error(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *tfe) forceCancelRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.client.ForceCancelRun(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// getPlan retrieves a plan object in JSON-API format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#show-a-plan
func (a *tfe) getPlan(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("plan_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	run, err := a.client.GetRun(r.Context(), resource.ConvertTfeID(id, "run"))
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	plan, err := a.toPlan(run.Plan, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, plan, http.StatusOK)
}

// getPlanJSON retrieves a plan object's plan file in JSON format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#retrieve-the-json-execution-plan
func (a *tfe) getPlanJSON(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("plan_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	json, err := a.client.GetRunPlanFile(r.Context(), resource.ConvertTfeID(id, "run"), PlanFormatJSON)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if _, err := w.Write(json); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *tfe) getApply(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("apply_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// otf's apply IDs are simply the corresponding run ID
	run, err := a.client.GetRun(r.Context(), resource.ConvertTfeID(id, "run"))
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	apply, err := a.toApply(run.Apply, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Responder.Respond(w, r, apply, http.StatusOK)
}

func (a *tfe) getLogs(w http.ResponseWriter, r *http.Request) {
	var opts GetChunkOptions
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	chunk, err := a.client.GetChunk(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if _, err := w.Write(chunk.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// OTF doesn't implement run events but as of terraform v1.5, the cloud backend
// makes a call to this endpoint. OTF therefore stubs this endpoint and sends an
// empty response, to avoid sending a 404 response and triggering an error.
func (a *tfe) listRunEvents(w http.ResponseWriter, r *http.Request) {
	a.Responder.Respond(w, r, []*TFERunEvent{}, http.StatusOK)
}

func (a *tfe) includeCurrentRun(ctx context.Context, v any) ([]any, error) {
	ws, ok := v.(*workspace.TFEWorkspace)
	if !ok {
		return nil, nil
	}
	if ws.CurrentRun == nil {
		return nil, nil
	}
	run, err := a.client.GetRun(ctx, ws.CurrentRun.ID)
	if err != nil {
		return nil, err
	}
	converted, err := a.toRun(run, ctx)
	if err != nil {
		return nil, err
	}
	return []any{converted}, nil
}

func (a *tfe) includeCreatedBy(ctx context.Context, v any) ([]any, error) {
	run, ok := v.(*TFERun)
	if !ok {
		return nil, nil
	}
	if run.CreatedBy == nil {
		return nil, nil
	}
	return []any{run.CreatedBy}, nil
}

func (a *tfe) includeWorkspace(ctx context.Context, v any) ([]any, error) {
	run, ok := v.(*TFERun)
	if !ok {
		return nil, nil
	}
	ws, err := a.client.GetWorkspace(ctx, run.Workspace.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace: %w", err)
	}
	include, err := workspace.ToTFE(a.authorizer, ws, (&http.Request{}).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return []any{include}, nil
}

// toRun converts a run into its equivalent json:api struct
func (a *tfe) toRun(from *Run, ctx context.Context) (*TFERun, error) {
	accessRequest := &authz.Request{ID: &from.ID}
	perms := &TFERunPermissions{
		CanDiscard:      a.authorizer.CanAccess(ctx, authz.DiscardRunAction, accessRequest),
		CanForceExecute: a.authorizer.CanAccess(ctx, authz.ApplyRunAction, accessRequest),
		CanForceCancel:  a.authorizer.CanAccess(ctx, authz.ForceCancelRunAction, accessRequest),
		CanCancel:       a.authorizer.CanAccess(ctx, authz.CancelRunAction, accessRequest),
		CanApply:        a.authorizer.CanAccess(ctx, authz.ApplyRunAction, accessRequest),
	}

	var timestamps TFERunStatusTimestamps
	for _, rst := range from.StatusTimestamps {
		switch rst.Status {
		case runstatus.Pending:
			timestamps.PlanQueueableAt = &rst.Timestamp
		case runstatus.PlanQueued:
			timestamps.PlanQueuedAt = &rst.Timestamp
		case runstatus.Planning:
			timestamps.PlanningAt = &rst.Timestamp
		case runstatus.PolicyChecking:
			timestamps.PolicyCheckingAt = &rst.Timestamp
		case runstatus.Planned:
			timestamps.PlannedAt = &rst.Timestamp
		case runstatus.CostEstimated:
			timestamps.CostEstimatedAt = &rst.Timestamp
		case runstatus.PolicyChecked:
			timestamps.PolicyCheckedAt = &rst.Timestamp
			if timestamps.CostEstimatedAt == nil {
				timestamps.CostEstimatedAt = &rst.Timestamp
			}
		case runstatus.PolicySoftFailed:
			timestamps.PolicySoftFailedAt = &rst.Timestamp
			if timestamps.CostEstimatedAt == nil {
				timestamps.CostEstimatedAt = &rst.Timestamp
			}
		case runstatus.PlannedAndFinished:
			timestamps.PlannedAndFinishedAt = &rst.Timestamp
		case runstatus.ApplyQueued:
			timestamps.ApplyQueuedAt = &rst.Timestamp
		case runstatus.Applying:
			timestamps.ApplyingAt = &rst.Timestamp
		case runstatus.Applied:
			timestamps.AppliedAt = &rst.Timestamp
		case runstatus.Errored:
			timestamps.ErroredAt = &rst.Timestamp
		case runstatus.Canceled:
			timestamps.CanceledAt = &rst.Timestamp
		case runstatus.ForceCanceled:
			timestamps.ForceCanceledAt = &rst.Timestamp
		case runstatus.Discarded:
			timestamps.DiscardedAt = &rst.Timestamp
		}
	}

	to := &TFERun{
		ID: from.ID,
		Actions: &TFERunActions{
			IsCancelable:      from.Cancelable(),
			IsConfirmable:     from.Confirmable(),
			IsForceCancelable: from.CancelSignaledAt != nil,
			IsDiscardable:     from.Discardable(),
		},
		AllowEmptyApply:  from.AllowEmptyApply,
		AutoApply:        from.AutoApply,
		CreatedAt:        from.CreatedAt,
		ExecutionMode:    string(from.ExecutionMode),
		HasChanges:       from.Plan.HasChanges(),
		IsDestroy:        from.IsDestroy,
		Message:          from.Message,
		Permissions:      perms,
		PlanOnly:         from.PlanOnly,
		PositionInQueue:  0,
		Refresh:          from.Refresh,
		RefreshOnly:      from.RefreshOnly,
		ReplaceAddrs:     from.ReplaceAddrs,
		Source:           string(from.Source),
		Status:           a.tfeStatus(from.Status),
		StatusTimestamps: &timestamps,
		TargetAddrs:      from.TargetAddrs,
		TerraformVersion: from.EngineVersion,
		// Relations
		Plan:     &TFEPlan{ID: resource.ConvertTfeID(from.ID, "plan")},
		Sentinel: &TFESentinel{ID: resource.ConvertTfeID(from.ID, "sentinel")},
		Apply:    &TFEApply{ID: resource.ConvertTfeID(from.ID, "apply")},
		// TODO: populate with real user.
		CreatedBy: &user.TFEUser{
			ID:       tfeUser,
			Username: "otf",
		},
		ConfigurationVersion: &configversion.TFEConfigurationVersion{
			ID: from.ConfigurationVersionID,
		},
		Workspace: &TFEWorkspace{ID: from.WorkspaceID},
	}
	to.Variables = make([]TFERunVariable, len(from.Variables))
	for i, from := range from.Variables {
		to.Variables[i] = TFERunVariable(from)
	}
	if from.CostEstimationEnabled {
		to.CostEstimate = &types.CostEstimate{ID: resource.ConvertTfeID(from.ID, "ce")}
	}
	//
	// go-tfe integration tests expect the ForceCancelAvailableAt parameter to
	// be set even if a run has already been successfully canceled gracefully
	// and its status has been set to RunCanceled; whereas OTF only permits a
	// run to be force canceled if the run has not been successfully canceled
	// and the run is yet to reach RunCanceled status. As a compromise, it is
	// set in either of these circumstances.
	if timestamps.CanceledAt != nil {
		// run successfully canceled
		cooledOff := timestamps.CanceledAt.Add(forceCancelCoolOff)
		to.ForceCancelAvailableAt = &cooledOff
	} else if from.CancelSignaledAt != nil {
		// run not successfully canceled yet
		cooledOff := from.CancelSignaledAt.Add(forceCancelCoolOff)
		to.ForceCancelAvailableAt = &cooledOff
	}
	return to, nil
}

func (a *tfe) tfeStatus(status runstatus.Status) string {
	switch status {
	case runstatus.PolicyChecked, runstatus.PolicySoftFailed:
		// Terraform Cloud clients still expect a completed pre-apply run to
		// surface as cost_estimated/planned-compatible even though OTF tracks
		// the internal Sentinel outcome separately.
		return string(runstatus.CostEstimated)
	default:
		return string(status)
	}
}

func (a *tfe) toPlan(plan Phase, r *http.Request) (*TFEPlan, error) {
	logURL, err := a.logURL(r, plan)
	if err != nil {
		return nil, err
	}

	return &TFEPlan{
		ID:                resource.ConvertTfeID(plan.RunID, "plan"),
		HasChanges:        plan.HasChanges(),
		LogReadURL:        logURL,
		TFEResourceReport: a.toResourceReport(plan.ResourceReport),
		Status:            string(plan.Status),
		StatusTimestamps:  a.toPhaseTimestamps(plan.StatusTimestamps),
	}, nil
}

func (a *tfe) toApply(apply Phase, r *http.Request) (*TFEApply, error) {
	logURL, err := a.logURL(r, apply)
	if err != nil {
		return nil, err
	}

	return &TFEApply{
		ID:                resource.ConvertTfeID(apply.RunID, "apply"),
		LogReadURL:        logURL,
		TFEResourceReport: a.toResourceReport(apply.ResourceReport),
		Status:            string(apply.Status),
		StatusTimestamps:  a.toPhaseTimestamps(apply.StatusTimestamps),
	}, nil
}

func (a *tfe) toResourceReport(from *Report) TFEResourceReport {
	var to TFEResourceReport
	if from != nil {
		to.Additions = &from.Additions
		to.Changes = &from.Changes
		to.Destructions = &from.Destructions
	}
	return to
}

func (a *tfe) toPhaseTimestamps(from []PhaseStatusTimestamp) *TFEPhaseStatusTimestamps {
	var timestamps TFEPhaseStatusTimestamps
	for _, ts := range from {
		switch ts.Status {
		case PhasePending:
			timestamps.PendingAt = &ts.Timestamp
		case PhaseCanceled:
			timestamps.CanceledAt = &ts.Timestamp
		case PhaseErrored:
			timestamps.ErroredAt = &ts.Timestamp
		case PhaseFinished:
			timestamps.FinishedAt = &ts.Timestamp
		case PhaseQueued:
			timestamps.QueuedAt = &ts.Timestamp
		case PhaseRunning:
			timestamps.StartedAt = &ts.Timestamp
		case PhaseUnreachable:
			timestamps.UnreachableAt = &ts.Timestamp
		}
	}
	return &timestamps
}

func (a *tfe) logURL(r *http.Request, phase Phase) (string, error) {
	logs := fmt.Sprintf("/runs/%s/logs/%s", phase.RunID, phase.PhaseType)
	logs, err := a.signer.Sign(logs, time.Now().Add(time.Hour))
	if err != nil {
		return "", err
	}
	// terraform CLI expects an absolute URL
	return otfhttp.Absolute(r, logs), nil
}
