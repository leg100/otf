package run

import (
	"net/http"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

// jsonapiMarshaler marshals workspace into a struct suitable for marshaling
// into json-api
type jsonapiMarshaler struct {
	*Run
	req *http.Request
	*handlers
}

func (r *jsonapiMarshaler) ToJSONAPI() any {
	subject, err := otf.SubjectFromContext(r.req.Context())
	if err != nil {
		panic(err.Error())
	}
	perms, err := r.ListWorkspacePermissions(r.req.Context(), r.WorkspaceID())
	if err != nil {
		panic(err.Error())
	}
	policy := &otf.WorkspacePolicy{
		Organization: r.Organization(),
		WorkspaceID:  r.WorkspaceID(),
		Permissions:  perms,
	}

	obj := &jsonapi.Run{
		ID: r.ID(),
		Actions: &jsonapi.RunActions{
			IsCancelable:      r.Cancelable(),
			IsConfirmable:     r.Confirmable(),
			IsForceCancelable: r.ForceCancelAvailableAt() != nil,
			IsDiscardable:     r.Discardable(),
		},
		CreatedAt:              r.CreatedAt(),
		ExecutionMode:          string(r.ExecutionMode()),
		ForceCancelAvailableAt: r.ForceCancelAvailableAt(),
		HasChanges:             r.Plan().HasChanges(),
		IsDestroy:              r.IsDestroy(),
		Message:                r.Message(),
		Permissions: &jsonapi.RunPermissions{
			CanDiscard:      subject.CanAccessWorkspace(rbac.DiscardRunAction, policy),
			CanForceExecute: subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
			CanForceCancel:  subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
			CanCancel:       subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
			CanApply:        subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		},
		PositionInQueue:  0,
		Refresh:          r.Refresh(),
		RefreshOnly:      r.RefreshOnly(),
		ReplaceAddrs:     r.ReplaceAddrs(),
		Source:           otf.DefaultConfigurationSource,
		Status:           string(r.Status()),
		StatusTimestamps: &jsonapi.RunStatusTimestamps{},
		TargetAddrs:      r.TargetAddrs(),
		// Relations
		Apply: (&apply{r.Apply(), r.req, r.Server}).ToJSONAPI().(*jsonapiApply),
		Plan:  (&plan{r.Plan(), r.req, r.Server}).ToJSONAPI().(*jsonapiPlan),
		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &jsonapi.User{
			ID:       otf.DefaultUserID,
			Username: otf.DefaultUsername,
		},
		ConfigurationVersion: &jsonapi.ConfigurationVersion{
			ID: r.ConfigurationVersionID(),
		},
		Workspace: &jsonapi.Workspace{ID: r.WorkspaceID()},
	}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#available-related-resources
	//
	// NOTE: limit support to workspace, since that's what the go-tfe tests
	// for, and we want to run the full barrage of go-tfe workspace tests
	// without error
	if includes := r.req.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "workspace":
				ws, err := r.Application.GetWorkspace(r.req.Context(), r.WorkspaceID())
				if err != nil {
					panic(err.Error()) // throws HTTP500
				}
				obj.Workspace = (&Workspace{r.req, r.Application, ws}).ToJSONAPI().(*jsonapi.Workspace)
			}
		}
	}

	for _, rst := range r.StatusTimestamps() {
		switch rst.Status {
		case otf.RunPending:
			obj.StatusTimestamps.PlanQueueableAt = &rst.Timestamp
		case otf.RunPlanQueued:
			obj.StatusTimestamps.PlanQueuedAt = &rst.Timestamp
		case otf.RunPlanning:
			obj.StatusTimestamps.PlanningAt = &rst.Timestamp
		case otf.RunPlanned:
			obj.StatusTimestamps.PlannedAt = &rst.Timestamp
		case otf.RunPlannedAndFinished:
			obj.StatusTimestamps.PlannedAndFinishedAt = &rst.Timestamp
		case otf.RunApplyQueued:
			obj.StatusTimestamps.ApplyQueuedAt = &rst.Timestamp
		case otf.RunApplying:
			obj.StatusTimestamps.ApplyingAt = &rst.Timestamp
		case otf.RunApplied:
			obj.StatusTimestamps.AppliedAt = &rst.Timestamp
		case otf.RunErrored:
			obj.StatusTimestamps.ErroredAt = &rst.Timestamp
		case otf.RunCanceled:
			obj.StatusTimestamps.CanceledAt = &rst.Timestamp
		case otf.RunForceCanceled:
			obj.StatusTimestamps.ForceCanceledAt = &rst.Timestamp
		case otf.RunDiscarded:
			obj.StatusTimestamps.DiscardedAt = &rst.Timestamp
		}
	}
	return obj
}

type jsonapiPlanMarshaler struct {
	*Plan
	req *http.Request
	*handlers
}

// ToJSONAPI assembles a JSON-API DTO.
func (p *jsonapiPlanMarshaler) ToJSONAPI() any {
	dto := &jsonapiPlan{
		ID:               otf.ConvertID(p.ID(), "plan"),
		HasChanges:       p.HasChanges(),
		LogReadURL:       p.signedLogURL(p.req, p.ID(), "plan"),
		Status:           string(p.Status()),
		StatusTimestamps: &jsonapiPhaseStatusTimestamps{},
	}
	if p.ResourceReport != nil {
		dto.Additions = &p.Additions
		dto.Changes = &p.Changes
		dto.Destructions = &p.Destructions
	}
	for _, ts := range p.StatusTimestamps() {
		switch ts.Status {
		case otf.PhasePending:
			dto.StatusTimestamps.PendingAt = &ts.Timestamp
		case otf.PhaseCanceled:
			dto.StatusTimestamps.CanceledAt = &ts.Timestamp
		case otf.PhaseErrored:
			dto.StatusTimestamps.ErroredAt = &ts.Timestamp
		case otf.PhaseFinished:
			dto.StatusTimestamps.FinishedAt = &ts.Timestamp
		case otf.PhaseQueued:
			dto.StatusTimestamps.QueuedAt = &ts.Timestamp
		case otf.PhaseRunning:
			dto.StatusTimestamps.StartedAt = &ts.Timestamp
		case otf.PhaseUnreachable:
			dto.StatusTimestamps.UnreachableAt = &ts.Timestamp
		}
	}
	return dto
}
