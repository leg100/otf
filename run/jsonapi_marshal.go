package run

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/workspace"
)

// JSONAPIMarshaler marshals structs into the json:api encoding
type JSONAPIMarshaler struct {
	workspace.Service // for retrieving workspace and workspace permissions

	*logURLGenerator
}

func newJSONAPIMarshaler(svc workspace.Service, signer otf.Signer) *JSONAPIMarshaler {
	return &JSONAPIMarshaler{
		Service:         svc,
		logURLGenerator: &logURLGenerator{signer},
	}
}

// MarshalJSONAPI marshals a run into json:api encoded data
func (m *JSONAPIMarshaler) MarshalJSONAPI(run *Run, r *http.Request) ([]byte, error) {
	jrun, err := m.toRun(run, r)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = jsonapi.MarshalPayloadWithoutIncluded(&buf, jrun); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// toRun converts a run into its equivalent json:api struct
func (m *JSONAPIMarshaler) toRun(run *Run, r *http.Request) (*jsonapi.Run, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	policy, err := m.GetPolicy(r.Context(), run.WorkspaceID)
	if err != nil {
		return nil, err
	}
	perms := &jsonapi.RunPermissions{
		CanDiscard:      subject.CanAccessWorkspace(rbac.DiscardRunAction, policy),
		CanForceExecute: subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanForceCancel:  subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanCancel:       subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanApply:        subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
	}

	workspace := &jsonapi.Workspace{ID: run.WorkspaceID}

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
				workspace, err = m.GetWorkspaceJSONAPI(r.Context(), run.WorkspaceID, r)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	var timestamps jsonapi.RunStatusTimestamps
	for _, rst := range run.StatusTimestamps {
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

	plan, err := m.toPlan(run.Plan, r)
	if err != nil {
		return nil, err
	}
	apply, err := m.toApply(run.Apply, r)
	if err != nil {
		return nil, err
	}

	return &jsonapi.Run{
		ID: run.ID,
		Actions: &jsonapi.RunActions{
			IsCancelable:      run.Cancelable(),
			IsConfirmable:     run.Confirmable(),
			IsForceCancelable: run.ForceCancelAvailableAt != nil,
			IsDiscardable:     run.Discardable(),
		},
		CreatedAt:              run.CreatedAt,
		ExecutionMode:          string(run.ExecutionMode),
		ForceCancelAvailableAt: run.ForceCancelAvailableAt,
		HasChanges:             run.Plan.HasChanges(),
		IsDestroy:              run.IsDestroy,
		Message:                run.Message,
		Permissions:            perms,
		PositionInQueue:        0,
		Refresh:                run.Refresh,
		RefreshOnly:            run.RefreshOnly,
		ReplaceAddrs:           run.ReplaceAddrs,
		Source:                 configversion.DefaultConfigurationSource,
		Status:                 string(run.Status),
		StatusTimestamps:       &timestamps,
		TargetAddrs:            run.TargetAddrs,
		// Relations
		Plan:  plan,
		Apply: apply,
		// Hardcoded anonymous user until authorization is introduced
		CreatedBy: &jsonapi.User{
			ID:       otf.DefaultUserID,
			Username: otf.DefaultUsername,
		},
		ConfigurationVersion: &jsonapi.ConfigurationVersion{
			ID: run.ConfigurationVersionID,
		},
		Workspace: workspace,
	}, nil
}

func (m JSONAPIMarshaler) toList(list *RunList, r *http.Request) (*jsonapi.RunList, error) {
	var items []*jsonapi.Run
	for _, run := range list.Items {
		jrun, err := m.toRun(run, r)
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

func (m *JSONAPIMarshaler) toPhase(from Phase, r *http.Request) (any, error) {
	switch from.PhaseType {
	case otf.PlanPhase:
		return m.toPlan(from, r)
	case otf.ApplyPhase:
		return m.toApply(from, r)
	default:
		return nil, fmt.Errorf("unsupported phase: %s", from.PhaseType)
	}
}

func (m *JSONAPIMarshaler) toPlan(plan Phase, r *http.Request) (*jsonapi.Plan, error) {
	logURL, err := m.logURL(r, plan)
	if err != nil {
		return nil, err
	}

	return &jsonapi.Plan{
		ID:               otf.ConvertID(plan.RunID, "plan"),
		HasChanges:       plan.HasChanges(),
		LogReadURL:       logURL,
		ResourceReport:   m.toResourceReport(plan.ResourceReport),
		Status:           string(plan.Status),
		StatusTimestamps: m.toPhaseTimestamps(plan.StatusTimestamps),
	}, nil
}

func (m *JSONAPIMarshaler) toApply(apply Phase, r *http.Request) (*jsonapi.Apply, error) {
	logURL, err := m.logURL(r, apply)
	if err != nil {
		return nil, err
	}

	return &jsonapi.Apply{
		ID:               otf.ConvertID(apply.RunID, "apply"),
		LogReadURL:       logURL,
		ResourceReport:   m.toResourceReport(apply.ResourceReport),
		Status:           string(apply.Status),
		StatusTimestamps: m.toPhaseTimestamps(apply.StatusTimestamps),
	}, nil
}

func (m *JSONAPIMarshaler) toResourceReport(from *ResourceReport) jsonapi.ResourceReport {
	var to jsonapi.ResourceReport
	if from != nil {
		to.Additions = &from.Additions
		to.Changes = &from.Changes
		to.Destructions = &from.Destructions
	}
	return to
}

func (m *JSONAPIMarshaler) toPhaseTimestamps(from []PhaseStatusTimestamp) *jsonapi.PhaseStatusTimestamps {
	var timestamps jsonapi.PhaseStatusTimestamps
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

// logURLGenerator creates a signed URL for retrieving logs for a run phase.
type logURLGenerator struct {
	otf.Signer
}

func (s *logURLGenerator) logURL(r *http.Request, phase Phase) (string, error) {
	logs := fmt.Sprintf("/runs/%s/logs/%s", phase.RunID, phase.PhaseType)
	logs, err := s.Sign(logs, time.Hour)
	if err != nil {
		return "", err
	}
	// Terraform CLI expects an absolute URL
	return otfhttp.Absolute(r, logs), nil
}
