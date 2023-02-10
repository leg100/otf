package run

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

// jsonapiMarshaler converts run into a struct suitable for
// marshaling into json:api encoding
type jsonapiMarshaler struct {
	otf.Signer      // for signing plan and apply log urls
	otf.Application // for retrieving workspace and workspace permissions
}

func (m *jsonapiMarshaler) toJSONAPI(run *Run, r *http.Request) (*jsonapi.Run, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, err
	}

	workspacePerms, err := m.ListWorkspacePermissions(r.Context(), run.WorkspaceID())
	if err != nil {
		return nil, err
	}
	policy := &otf.WorkspacePolicy{
		Organization: run.Organization(),
		WorkspaceID:  run.WorkspaceID(),
		Permissions:  workspacePerms,
	}

	runPerms := &jsonapi.RunPermissions{
		CanDiscard:      subject.CanAccessWorkspace(rbac.DiscardRunAction, policy),
		CanForceExecute: subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanForceCancel:  subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanCancel:       subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanApply:        subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
	}

	workspace := &jsonapi.Workspace{ID: run.WorkspaceID()}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#available-related-resources
	//
	// NOTE: limit support to workspace, since that's what the go-tfe tests
	// for, and we want to run the full barrage of go-tfe workspace tests
	// without error
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "workspace":
				workspace, err = m.GetWorkspaceJSONAPI(r.Context(), run.WorkspaceID())
				if err != nil {
					return nil, err
				}
			}
		}
	}

	var timestamps jsonapi.RunStatusTimestamps
	for _, rst := range run.StatusTimestamps() {
		switch rst.Status {
		case otf.RunPending:
			timestamps.PlanQueueableAt = &rst.Timestamp
		case otf.RunPlanQueued:
			timestamps.PlanQueuedAt = &rst.Timestamp
		case otf.RunPlanning:
			timestamps.PlanningAt = &rst.Timestamp
		case otf.RunPlanned:
			timestamps.PlannedAt = &rst.Timestamp
		case otf.RunPlannedAndFinished:
			timestamps.PlannedAndFinishedAt = &rst.Timestamp
		case otf.RunApplyQueued:
			timestamps.ApplyQueuedAt = &rst.Timestamp
		case otf.RunApplying:
			timestamps.ApplyingAt = &rst.Timestamp
		case otf.RunApplied:
			timestamps.AppliedAt = &rst.Timestamp
		case otf.RunErrored:
			timestamps.ErroredAt = &rst.Timestamp
		case otf.RunCanceled:
			timestamps.CanceledAt = &rst.Timestamp
		case otf.RunForceCanceled:
			timestamps.ForceCanceledAt = &rst.Timestamp
		case otf.RunDiscarded:
			timestamps.DiscardedAt = &rst.Timestamp
		}
	}

	plan, err := m.planToJSONAPI(run.plan, r)
	if err != nil {
		return nil, err
	}
	apply, err := m.applyToJSONAPI(run.apply, r)
	if err != nil {
		return nil, err
	}

	return &jsonapi.Run{
		ID: run.ID(),
		Actions: &jsonapi.RunActions{
			IsCancelable:      run.Cancelable(),
			IsConfirmable:     run.Confirmable(),
			IsForceCancelable: run.ForceCancelAvailableAt() != nil,
			IsDiscardable:     run.Discardable(),
		},
		CreatedAt:              run.CreatedAt(),
		ExecutionMode:          string(run.ExecutionMode()),
		ForceCancelAvailableAt: run.ForceCancelAvailableAt(),
		HasChanges:             run.Plan().HasChanges(),
		IsDestroy:              run.IsDestroy(),
		Message:                run.Message(),
		Permissions:            runPerms,
		PositionInQueue:        0,
		Refresh:                run.Refresh(),
		RefreshOnly:            run.RefreshOnly(),
		ReplaceAddrs:           run.ReplaceAddrs(),
		Source:                 otf.DefaultConfigurationSource,
		Status:                 string(run.Status()),
		StatusTimestamps:       &timestamps,
		TargetAddrs:            run.TargetAddrs(),
		// Relations
		Plan:  plan,
		Apply: apply,
		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &jsonapi.User{
			ID:       otf.DefaultUserID,
			Username: otf.DefaultUsername,
		},
		ConfigurationVersion: &jsonapi.ConfigurationVersion{
			ID: run.ConfigurationVersionID(),
		},
		Workspace: workspace,
	}, nil
}

type jsonapiPlanMarshaler struct {
	*Plan
	req *http.Request
	*handlers
}

// ToJSONAPI assembles a JSON-API DTO.
func (m *jsonapiMarshaler) planToJSONAPI(plan *Plan, r *http.Request) (*jsonapi.Plan, error) {
	var report *jsonapi.ResourceReport
	if plan.ResourceReport != nil {
		report.Additions = &plan.Additions
		report.Changes = &plan.Changes
		report.Destructions = &plan.Destructions
	}

	var timestamps jsonapi.PhaseStatusTimestamps
	for _, ts := range plan.StatusTimestamps() {
		switch ts.Status {
		case otf.PhasePending:
			timestamps.PendingAt = &ts.Timestamp
		case otf.PhaseCanceled:
			timestamps.CanceledAt = &ts.Timestamp
		case otf.PhaseErrored:
			timestamps.ErroredAt = &ts.Timestamp
		case otf.PhaseFinished:
			timestamps.FinishedAt = &ts.Timestamp
		case otf.PhaseQueued:
			timestamps.QueuedAt = &ts.Timestamp
		case otf.PhaseRunning:
			timestamps.StartedAt = &ts.Timestamp
		case otf.PhaseUnreachable:
			timestamps.UnreachableAt = &ts.Timestamp
		}
	}

	// signedLogURL creates a signed URL for retrieving logs for a run phase.
	logs := fmt.Sprintf("/runs/%s/logs/%s", plan.runID, otf.PlanPhase)
	logs, err := m.Sign(logs, time.Hour)
	if err != nil {
		return nil, err
	}
	// Terraform CLI expects an absolute URL
	logs = otfhttp.Absolute(r, logs)

	return &jsonapi.Plan{
		ID:               otf.ConvertID(plan.ID(), "plan"),
		HasChanges:       plan.HasChanges(),
		LogReadURL:       logs,
		Status:           string(plan.Status()),
		StatusTimestamps: &timestamps,
	}, nil
}

func (m *jsonapiMarshaler) applyToJSONAPI(apply *Apply, r *http.Request) (*jsonapi.Apply, error) {
	var report *jsonapi.ResourceReport
	if apply.ResourceReport != nil {
		report.Additions = &apply.Additions
		report.Changes = &apply.Changes
		report.Destructions = &apply.Destructions
	}

	var timestamps jsonapi.PhaseStatusTimestamps
	for _, ts := range apply.StatusTimestamps() {
		switch ts.Status {
		case otf.PhasePending:
			timestamps.PendingAt = &ts.Timestamp
		case otf.PhaseCanceled:
			timestamps.CanceledAt = &ts.Timestamp
		case otf.PhaseErrored:
			timestamps.ErroredAt = &ts.Timestamp
		case otf.PhaseFinished:
			timestamps.FinishedAt = &ts.Timestamp
		case otf.PhaseQueued:
			timestamps.QueuedAt = &ts.Timestamp
		case otf.PhaseRunning:
			timestamps.StartedAt = &ts.Timestamp
		case otf.PhaseUnreachable:
			timestamps.UnreachableAt = &ts.Timestamp
		}
	}

	// signedLogURL creates a signed URL for retrieving logs for a run phase.
	logs := fmt.Sprintf("/runs/%s/logs/%s", apply.runID, otf.ApplyPhase)
	logs, err := m.Sign(logs, time.Hour)
	if err != nil {
		return nil, err
	}
	// Terraform CLI expects an absolute URL
	logs = otfhttp.Absolute(r, logs)

	return &jsonapi.Apply{
		ID:               otf.ConvertID(apply.ID(), "apply"),
		LogReadURL:       logs,
		Status:           string(apply.Status()),
		StatusTimestamps: &timestamps,
	}, nil
}

func (m jsonapiMarshaler) toJSONAPIList(list *RunList, r *http.Request) (*jsonapi.RunList, error) {
	var items []*jsonapi.Run
	for _, run := range list.Items {
		jrun, err := m.toJSONAPI(run, r)
		if err != nil {
			return nil, err
		}
		items = append(items, jrun)
	}
	return &jsonapi.RunList{
		Items:      items,
		Pagination: list.Pagination.ToJSONAPI(),
	}, nil
}
