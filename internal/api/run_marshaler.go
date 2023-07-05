package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/configversion"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/run"
)

// runLogsURLGenerator creates a signed URL for retrieving logs for a run phase.
type runLogsURLGenerator struct {
	internal.Signer
}

// toRun converts a run into its equivalent json:api struct
func (m *jsonapiMarshaler) toRun(from *run.Run, r *http.Request) (*types.Run, []jsonapi.MarshalOption, error) {
	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		return nil, nil, err
	}
	policy, err := m.GetPolicy(r.Context(), from.WorkspaceID)
	if err != nil {
		return nil, nil, err
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
		case internal.RunPending:
			timestamps.PlanQueueableAt = &rst.Timestamp
		case internal.RunPlanQueued:
			timestamps.PlanQueuedAt = &rst.Timestamp
		case internal.RunPlanning:
			timestamps.PlanningAt = &rst.Timestamp
		case internal.RunPlanned:
			timestamps.PlannedAt = &rst.Timestamp
		case internal.RunPlannedAndFinished:
			timestamps.PlannedAndFinishedAt = &rst.Timestamp
		case internal.RunApplyQueued:
			timestamps.ApplyQueuedAt = &rst.Timestamp
		case internal.RunApplying:
			timestamps.ApplyingAt = &rst.Timestamp
		case internal.RunApplied:
			timestamps.AppliedAt = &rst.Timestamp
		case internal.RunErrored:
			timestamps.ErroredAt = &rst.Timestamp
		case internal.RunCanceled:
			timestamps.CanceledAt = &rst.Timestamp
		case internal.RunForceCanceled:
			timestamps.ForceCanceledAt = &rst.Timestamp
		case internal.RunDiscarded:
			timestamps.DiscardedAt = &rst.Timestamp
		}
	}

	plan, err := m.toPlan(from.Plan, r)
	if err != nil {
		return nil, nil, err
	}
	apply, err := m.toApply(from.Apply, r)
	if err != nil {
		return nil, nil, err
	}

	to := &types.Run{
		ID: from.ID,
		Actions: &types.RunActions{
			IsCancelable:      from.Cancelable(),
			IsConfirmable:     from.Confirmable(),
			IsForceCancelable: from.ForceCancelAvailableAt != nil,
			IsDiscardable:     from.Discardable(),
		},
		CreatedAt:              from.CreatedAt,
		ExecutionMode:          string(from.ExecutionMode),
		ForceCancelAvailableAt: from.ForceCancelAvailableAt,
		HasChanges:             from.Plan.HasChanges(),
		IsDestroy:              from.IsDestroy,
		Message:                from.Message,
		Permissions:            perms,
		PositionInQueue:        0,
		Refresh:                from.Refresh,
		RefreshOnly:            from.RefreshOnly,
		ReplaceAddrs:           from.ReplaceAddrs,
		Source:                 configversion.DefaultConfigurationSource,
		Status:                 string(from.Status),
		StatusTimestamps:       &timestamps,
		TargetAddrs:            from.TargetAddrs,
		// Relations
		Plan:  plan,
		Apply: apply,
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
				unmarshaled, err := m.GetWorkspace(r.Context(), from.WorkspaceID)
				if err != nil {
					return nil, nil, err
				}
				workspace, _, err := m.toWorkspace(unmarshaled, r)
				if err != nil {
					return nil, nil, err
				}
				opts = append(opts, jsonapi.MarshalInclude(workspace))
			case "created_by":
				opts = append(opts, jsonapi.MarshalInclude(to.CreatedBy))
			}
		}
	}

	return to, opts, nil
}

func (m jsonapiMarshaler) toRunList(from *run.RunList, r *http.Request) (to []*types.Run, opts []jsonapi.MarshalOption, err error) {
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
	case internal.PlanPhase:
		return m.toPlan(from, r)
	case internal.ApplyPhase:
		return m.toApply(from, r)
	default:
		return nil, fmt.Errorf("unsupported phase: %s", from.PhaseType)
	}
}

func (m *jsonapiMarshaler) toPlan(plan run.Phase, r *http.Request) (*types.Plan, error) {
	logURL, err := m.logURL(r, plan)
	if err != nil {
		return nil, err
	}

	return &types.Plan{
		ID:               internal.ConvertID(plan.RunID, "plan"),
		HasChanges:       plan.HasChanges(),
		LogReadURL:       logURL,
		ResourceReport:   m.toResourceReport(plan.ResourceReport),
		Status:           string(plan.Status),
		StatusTimestamps: m.toPhaseTimestamps(plan.StatusTimestamps),
	}, nil
}

func (m *jsonapiMarshaler) toApply(apply run.Phase, r *http.Request) (*types.Apply, error) {
	logURL, err := m.logURL(r, apply)
	if err != nil {
		return nil, err
	}

	return &types.Apply{
		ID:               internal.ConvertID(apply.RunID, "apply"),
		LogReadURL:       logURL,
		ResourceReport:   m.toResourceReport(apply.ResourceReport),
		Status:           string(apply.Status),
		StatusTimestamps: m.toPhaseTimestamps(apply.StatusTimestamps),
	}, nil
}

func (m *jsonapiMarshaler) toResourceReport(from *run.Report) types.ResourceReport {
	var to types.ResourceReport
	if from != nil {
		to.Additions = &from.Additions
		to.Changes = &from.Changes
		to.Destructions = &from.Destructions
	}
	return to
}

func (m *jsonapiMarshaler) toPhaseTimestamps(from []run.PhaseStatusTimestamp) *types.PhaseStatusTimestamps {
	var timestamps types.PhaseStatusTimestamps
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
