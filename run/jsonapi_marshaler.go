package run

import (
	"net/http"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

// jsonapiMarshaler converts run into a struct suitable for
// marshaling into json:api encoding
type jsonapiMarshaler struct {
	otf.Signer      // for signing upload url
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
				ws, err := m.GetWorkspace(r.Context(), run.WorkspaceID())
				if err != nil {
					return marshalable{}, err
				}
				workspace = (&Workspace{r.req, r.Application, ws}).ToJSONAPI().(*jsonapi.Workspace)
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
		Apply: (&apply{run.Apply(), r.req, r.Server}).ToJSONAPI().(*jsonapiApply),
		Plan:  (&plan{run.Plan(), r.req, r.Server}).ToJSONAPI().(*jsonapiPlan),
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

type apply struct {
	*otf.Apply
	req *http.Request
	*handlers
}

type RunList struct {
	*otf.RunList
	req *http.Request
	*handlers
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *RunList) ToJSONAPI() any {
	obj := &jsonapi.RunList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Run{item, l.req, l.Server}).ToJSONAPI().(*jsonapi.Run))
	}
	return obj
}
