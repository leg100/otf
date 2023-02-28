package run

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

// JSONAPIConverter converts a run into a json:api struct
type JSONAPIConverter struct {
	otf.WorkspaceService // for retrieving workspace and workspace permissions

	*jsonapiPlanConverter
	*jsonapiApplyConverter
}

func newJSONAPIConverter(svc otf.WorkspaceService, signer otf.Signer) *JSONAPIConverter {
	return &JSONAPIConverter{
		WorkspaceService: svc,
		jsonapiPlanConverter: &jsonapiPlanConverter{
			logURLGenerator: &logURLGenerator{signer, otf.PlanPhase},
		},
		jsonapiApplyConverter: &jsonapiApplyConverter{
			logURLGenerator: &logURLGenerator{signer, otf.ApplyPhase},
		},
	}
}

// MarshalJSONAPI marshals a run into json:api encoded data
func (m *JSONAPIConverter) MarshalJSONAPI(run *Run, r *http.Request) ([]byte, error) {
	jrun, err := m.toJSONAPI(run, r)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = jsonapi.MarshalPayloadWithoutIncluded(&buf, jrun); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *JSONAPIConverter) toJSONAPI(run *Run, r *http.Request) (*jsonapi.Run, error) {
	subject, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		return nil, err
	}
	policy, err := m.GetPolicy(r.Context(), run.WorkspaceID())
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

	plan, err := m.plan().toJSONAPI(run.plan, r)
	if err != nil {
		return nil, err
	}
	apply, err := m.apply().toJSONAPI(run.apply, r)
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
		Permissions:            perms,
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

func (m JSONAPIConverter) toJSONAPIList(list *RunList, r *http.Request) (*jsonapi.RunList, error) {
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

func (m *JSONAPIConverter) plan() *jsonapiPlanConverter   { return m.jsonapiPlanConverter }
func (m *JSONAPIConverter) apply() *jsonapiApplyConverter { return m.jsonapiApplyConverter }

// jsonapiPlanConverter converts a plan into a json:api struct
type jsonapiPlanConverter struct {
	*logURLGenerator
}

func (m *jsonapiPlanConverter) toJSONAPI(plan *Plan, r *http.Request) (*jsonapi.Plan, error) {
	var report jsonapi.ResourceReport
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

	logURL, err := m.logURL(r, plan.runID)
	if err != nil {
		return nil, err
	}

	return &jsonapi.Plan{
		ID:               otf.ConvertID(plan.ID(), "plan"),
		HasChanges:       plan.HasChanges(),
		LogReadURL:       logURL,
		ResourceReport:   report,
		Status:           string(plan.Status()),
		StatusTimestamps: &timestamps,
	}, nil
}

// jsonapiApplyConverter converts an apply into a json:api struct
type jsonapiApplyConverter struct {
	*logURLGenerator
}

func (m *jsonapiApplyConverter) toJSONAPI(apply *Apply, r *http.Request) (*jsonapi.Apply, error) {
	var report jsonapi.ResourceReport
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

	logURL, err := m.logURL(r, apply.runID)
	if err != nil {
		return nil, err
	}

	return &jsonapi.Apply{
		ID:               otf.ConvertID(apply.ID(), "apply"),
		LogReadURL:       logURL,
		ResourceReport:   report,
		Status:           string(apply.Status()),
		StatusTimestamps: &timestamps,
	}, nil
}

// logURLGenerator creates a signed URL for retrieving logs for a run phase.
type logURLGenerator struct {
	otf.Signer

	phase otf.PhaseType
}

func (s *logURLGenerator) logURL(r *http.Request, runID string) (string, error) {
	logs := fmt.Sprintf("/runs/%s/logs/%s", runID, s.phase)
	logs, err := s.Sign(logs, time.Hour)
	if err != nil {
		return "", err
	}
	// Terraform CLI expects an absolute URL
	return otfhttp.Absolute(r, logs), nil
}
