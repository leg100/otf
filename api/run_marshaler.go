package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/run"
)

// runLogsURLGenerator creates a signed URL for retrieving logs for a run phase.
type runLogsURLGenerator struct {
	otf.Signer
}

// toRun converts a run into its equivalent json:api struct
func (m *jsonapiMarshaler) toRun(run *run.Run, r *http.Request) (*Run, []jsonapi.MarshalOption, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, nil, err
	}
	policy, err := m.GetPolicy(r.Context(), run.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	perms := &RunPermissions{
		CanDiscard:      subject.CanAccessWorkspace(rbac.DiscardRunAction, policy),
		CanForceExecute: subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
		CanForceCancel:  subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanCancel:       subject.CanAccessWorkspace(rbac.CancelRunAction, policy),
		CanApply:        subject.CanAccessWorkspace(rbac.ApplyRunAction, policy),
	}

	workspace := &Workspace{ID: run.WorkspaceID}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#available-related-resources
	//
	// NOTE: limit support to workspace, since that's what the go-tfe tests
	// for, and we want to run the full barrage of go-tfe workspace tests
	// without error
	var opts []jsonapi.MarshalOption
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "workspace":
				unmarshaled, err := m.GetWorkspace(r.Context(), run.WorkspaceID)
				if err != nil {
					return nil, nil, err
				}
				workspace, _, err = m.toWorkspace(unmarshaled, r)
				if err != nil {
					return nil, nil, err
				}
				opts = append(opts, jsonapi.MarshalInclude(workspace))
			}
		}
	}

	var timestamps RunStatusTimestamps
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
		return nil, nil, err
	}
	apply, err := m.toApply(run.Apply, r)
	if err != nil {
		return nil, nil, err
	}

	return &Run{
		ID: run.ID,
		Actions: &RunActions{
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
		// TODO: populate with real user.
		CreatedBy: &User{
			ID:       "user-123",
			Username: "otf",
		},
		ConfigurationVersion: &ConfigurationVersion{
			ID: run.ConfigurationVersionID,
		},
		Workspace: workspace,
	}, opts, nil
}

func (m jsonapiMarshaler) toRunList(from *run.RunList, r *http.Request) (to []*Run, opts []jsonapi.MarshalOption, err error) {
	opts = []jsonapi.MarshalOption{toMarshalOption(from.Pagination)}
	for _, i := range from.Items {
		run, itemOpts, err := m.toRun(i, r)
		if err != nil {
			return nil, nil, err
		}
		to = append(to, run)
		opts = append(opts, itemOpts...)
	}
	return to, opts, nil
}

func (m *jsonapiMarshaler) toPhase(from run.Phase, r *http.Request) (any, error) {
	switch from.PhaseType {
	case otf.PlanPhase:
		return m.toPlan(from, r)
	case otf.ApplyPhase:
		return m.toApply(from, r)
	default:
		return nil, fmt.Errorf("unsupported phase: %s", from.PhaseType)
	}
}

func (m *jsonapiMarshaler) toPlan(plan run.Phase, r *http.Request) (*Plan, error) {
	logURL, err := m.logURL(r, plan)
	if err != nil {
		return nil, err
	}

	return &Plan{
		ID:               otf.ConvertID(plan.RunID, "plan"),
		HasChanges:       plan.HasChanges(),
		LogReadURL:       logURL,
		ResourceReport:   m.toResourceReport(plan.ResourceReport),
		Status:           string(plan.Status),
		StatusTimestamps: m.toPhaseTimestamps(plan.StatusTimestamps),
	}, nil
}

func (m *jsonapiMarshaler) toApply(apply run.Phase, r *http.Request) (*Apply, error) {
	logURL, err := m.logURL(r, apply)
	if err != nil {
		return nil, err
	}

	return &Apply{
		ID:               otf.ConvertID(apply.RunID, "apply"),
		LogReadURL:       logURL,
		ResourceReport:   m.toResourceReport(apply.ResourceReport),
		Status:           string(apply.Status),
		StatusTimestamps: m.toPhaseTimestamps(apply.StatusTimestamps),
	}, nil
}

func (m *jsonapiMarshaler) toResourceReport(from *run.ResourceReport) ResourceReport {
	var to ResourceReport
	if from != nil {
		to.Additions = &from.Additions
		to.Changes = &from.Changes
		to.Destructions = &from.Destructions
	}
	return to
}

func (m *jsonapiMarshaler) toPhaseTimestamps(from []run.PhaseStatusTimestamp) *PhaseStatusTimestamps {
	var timestamps PhaseStatusTimestamps
	for _, ts := range from {
		switch ts.Status {
		case run.PhasePending:
			timestamps.PendingAt = &ts.Timestamp
		case run.PhaseCanceled:
			timestamps.CanceledAt = &ts.Timestamp
		case run.PhaseErrored:
			timestamps.ErroredAt = &ts.Timestamp
		case run.PhaseFinished:
			timestamps.FinishedAt = &ts.Timestamp
		case run.PhaseQueued:
			timestamps.QueuedAt = &ts.Timestamp
		case run.PhaseRunning:
			timestamps.StartedAt = &ts.Timestamp
		case run.PhaseUnreachable:
			timestamps.UnreachableAt = &ts.Timestamp
		}
	}
	return &timestamps
}

func (s *runLogsURLGenerator) logURL(r *http.Request, phase run.Phase) (string, error) {
	logs := fmt.Sprintf("/runs/%s/logs/%s", phase.RunID, phase.PhaseType)
	logs, err := s.Sign(logs, time.Hour)
	if err != nil {
		return "", err
	}
	// Terraform CLI expects an absolute URL
	return otfhttp.Absolute(r, logs), nil
}
