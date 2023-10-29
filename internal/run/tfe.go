package run

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
)

type tfe struct {
	Service
	workspace.PermissionsService
	internal.Signer
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

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
}

func (a *tfe) createRun(w http.ResponseWriter, r *http.Request) {
	var params types.RunCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if params.Workspace == nil {
		tfeapi.Error(w, &internal.MissingParameterError{Parameter: "workspace"})
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
		Source:           SourceAPI,
		AllowEmptyApply:  params.AllowEmptyApply,
		TerraformVersion: params.TerraformVersion,
	}
	if params.ConfigurationVersion != nil {
		opts.ConfigurationVersionID = &params.ConfigurationVersion.ID
	}
	if tfeapi.IsTerraformCLI(r) {
		opts.Source = SourceTerraform
	}
	opts.Variables = make([]Variable, len(params.Variables))
	for i, from := range params.Variables {
		opts.Variables[i] = Variable{Key: from.Key, Value: from.Value}
	}

	run, err := a.CreateRun(r.Context(), params.Workspace.ID, opts)
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
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	run, err := a.GetRun(r.Context(), id)
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
	var params types.RunListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert comma-separated list of statuses to []RunStatus
	statuses := internal.FromStringCSV[Status](params.Status)
	// convert comma-separated list of sources to []RunSource
	sources := internal.FromStringCSV[Source](params.Source)
	// split operations CSV
	operations := internal.SplitCSV(params.Operation)
	var planOnly *bool
	if slices.Contains(operations, string(types.RunOperationPlanOnly)) {
		planOnly = internal.Bool(true)
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
		Statuses: []Status{RunPlanQueued, RunApplyQueued},
	})
}

func (a *tfe) listRunsWithOptions(w http.ResponseWriter, r *http.Request, opts ListOptions) {
	page, err := a.ListRuns(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*types.Run, len(page.Items))
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
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Apply(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *tfe) discardRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err = a.DiscardRun(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *tfe) cancelRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if _, err = a.Cancel(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *tfe) forceCancelRun(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.ForceCancelRun(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// getPlan retrieves a plan object in JSON-API format.
//
// https://www.terraform.io/cloud-docs/api-docs/plans#show-a-plan
func (a *tfe) getPlan(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("plan_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	run, err := a.GetRun(r.Context(), internal.ConvertID(id, "run"))
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
	id, err := decode.Param("plan_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// otf's plan IDs are simply the corresponding run ID
	json, err := a.GetPlanFile(r.Context(), internal.ConvertID(id, "run"), PlanFormatJSON)
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
	id, err := decode.Param("apply_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// otf's apply IDs are simply the corresponding run ID
	run, err := a.GetRun(r.Context(), internal.ConvertID(id, "run"))
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	apply, err := a.toApply(run.Apply, r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, apply, http.StatusOK)
}

// OTF doesn't implement run events but as of terraform v1.5, the cloud backend
// makes a call to this endpoint. OTF therefore stubs this endpoint and sends an
// empty response, to avoid sending a 404 response and triggering an error.
func (a *tfe) listRunEvents(w http.ResponseWriter, r *http.Request) {
	a.Respond(w, r, []*types.RunEvent{}, http.StatusOK)
}

func (a *tfe) includeCurrentRun(ctx context.Context, v any) ([]any, error) {
	ws, ok := v.(*types.Workspace)
	if !ok {
		return nil, nil
	}
	if ws.CurrentRun == nil {
		return nil, nil
	}
	run, err := a.GetRun(ctx, ws.CurrentRun.ID)
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
	run, ok := v.(*types.Run)
	if !ok {
		return nil, nil
	}
	if run.CreatedBy == nil {
		return nil, nil
	}
	return []any{run.CreatedBy}, nil
}

// toRun converts a run into its equivalent json:api struct
func (a *tfe) toRun(from *Run, ctx context.Context) (*types.Run, error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	policy, err := a.GetPolicy(ctx, from.WorkspaceID)
	if err != nil {
		return nil, err
	}
	perms := &types.RunPermissions{
		CanDiscard:      subject.CanAccessWorkspace(rbac.DiscardRunAction, policy),
		CanForceExecute: subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanForceCancel:  subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanCancel:       subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanApply:        subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
	}

	var timestamps types.RunStatusTimestamps
	for _, rst := range from.StatusTimestamps {
		switch rst.Status {
		case RunPending:
			timestamps.PlanQueueableAt = &rst.Timestamp
		case RunPlanQueued:
			timestamps.PlanQueuedAt = &rst.Timestamp
		case RunPlanning:
			timestamps.PlanningAt = &rst.Timestamp
		case RunPlanned:
			timestamps.PlannedAt = &rst.Timestamp
		case RunPlannedAndFinished:
			timestamps.PlannedAndFinishedAt = &rst.Timestamp
		case RunApplyQueued:
			timestamps.ApplyQueuedAt = &rst.Timestamp
		case RunApplying:
			timestamps.ApplyingAt = &rst.Timestamp
		case RunApplied:
			timestamps.AppliedAt = &rst.Timestamp
		case RunErrored:
			timestamps.ErroredAt = &rst.Timestamp
		case RunCanceled:
			timestamps.CanceledAt = &rst.Timestamp
		case RunForceCanceled:
			timestamps.ForceCanceledAt = &rst.Timestamp
		case RunDiscarded:
			timestamps.DiscardedAt = &rst.Timestamp
		}
	}

	to := &types.Run{
		ID: from.ID,
		Actions: &types.RunActions{
			IsCancelable:      from.Cancelable(),
			IsConfirmable:     from.Confirmable(),
			IsForceCancelable: from.ForceCancelAvailableAt != nil,
			IsDiscardable:     from.Discardable(),
		},
		AllowEmptyApply:        from.AllowEmptyApply,
		AutoApply:              from.AutoApply,
		CreatedAt:              from.CreatedAt,
		ExecutionMode:          string(from.ExecutionMode),
		ForceCancelAvailableAt: from.ForceCancelAvailableAt,
		HasChanges:             from.Plan.HasChanges(),
		IsDestroy:              from.IsDestroy,
		Message:                from.Message,
		Permissions:            perms,
		PlanOnly:               from.PlanOnly,
		PositionInQueue:        0,
		Refresh:                from.Refresh,
		RefreshOnly:            from.RefreshOnly,
		ReplaceAddrs:           from.ReplaceAddrs,
		Source:                 string(from.Source),
		Status:                 string(from.Status),
		StatusTimestamps:       &timestamps,
		TargetAddrs:            from.TargetAddrs,
		TerraformVersion:       from.TerraformVersion,
		// Relations
		Plan:  &types.Plan{ID: internal.ConvertID(from.ID, "plan")},
		Apply: &types.Apply{ID: internal.ConvertID(from.ID, "apply")},
		// TODO: populate with real user.
		CreatedBy: &types.User{
			ID:       "user-123",
			Username: "otf",
		},
		ConfigurationVersion: &types.ConfigurationVersion{
			ID: from.ConfigurationVersionID,
		},
		Workspace: &types.Workspace{ID: from.WorkspaceID},
	}
	to.Variables = make([]types.RunVariable, len(from.Variables))
	for i, from := range from.Variables {
		to.Variables[i] = types.RunVariable{Key: from.Key, Value: from.Value}
	}
	if from.CostEstimationEnabled {
		to.CostEstimate = &types.CostEstimate{ID: internal.ConvertID(from.ID, "ce")}
	}

	return to, nil
}

func (a *tfe) toPlan(plan Phase, r *http.Request) (*types.Plan, error) {
	logURL, err := a.logURL(r, plan)
	if err != nil {
		return nil, err
	}

	return &types.Plan{
		ID:               internal.ConvertID(plan.RunID, "plan"),
		HasChanges:       plan.HasChanges(),
		LogReadURL:       logURL,
		ResourceReport:   a.toResourceReport(plan.ResourceReport),
		Status:           string(plan.Status),
		StatusTimestamps: a.toPhaseTimestamps(plan.StatusTimestamps),
	}, nil
}

func (a *tfe) toApply(apply Phase, r *http.Request) (*types.Apply, error) {
	logURL, err := a.logURL(r, apply)
	if err != nil {
		return nil, err
	}

	return &types.Apply{
		ID:               internal.ConvertID(apply.RunID, "apply"),
		LogReadURL:       logURL,
		ResourceReport:   a.toResourceReport(apply.ResourceReport),
		Status:           string(apply.Status),
		StatusTimestamps: a.toPhaseTimestamps(apply.StatusTimestamps),
	}, nil
}

func (a *tfe) toResourceReport(from *Report) types.ResourceReport {
	var to types.ResourceReport
	if from != nil {
		to.Additions = &from.Additions
		to.Changes = &from.Changes
		to.Destructions = &from.Destructions
	}
	return to
}

func (a *tfe) toPhaseTimestamps(from []PhaseStatusTimestamp) *types.PhaseStatusTimestamps {
	var timestamps types.PhaseStatusTimestamps
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
	logs, err := a.Sign(logs, time.Hour)
	if err != nil {
		return "", err
	}
	// terraform CLI expects an absolute URL
	return otfhttp.Absolute(r, logs), nil
}
