package api

import (
	"fmt"
	"net/http"

	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/workspace"
)

type (
	marshaler interface {
		toRun(run *run.Run, r *http.Request) (*jsonapi.Run, error)
		toRunList(from *run.RunList, r *http.Request) (*jsonapi.RunList, error)
		toPhase(from run.Phase, r *http.Request) (any, error)
		toPlan(plan run.Phase, r *http.Request) (*jsonapi.Plan, error)
		writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter))
	}

	jsonapiMarshaler struct {
		run.RunService
		organization.OrganizationService
		state.StateService
		workspace.WorkspaceService

		*runLogsURLGenerator
	}
)

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (m *jsonapiMarshaler) writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter)) {
	var (
		payload any
		err     error
	)

	switch v := v.(type) {
	case *organization.OrganizationList:
		payload = m.toOrganizationList(v)
	case *organization.Organization:
		payload = m.toOrganization(v)
	case organization.Entitlements:
		payload = (*jsonapi.Entitlements)(&v)
	case *workspace.WorkspaceList:
		payload, err = m.toWorkspaceList(v, r)
	case *workspace.Workspace:
		payload, err = m.toWorkspace(v, r)
	case *run.RunList:
		payload, err = m.toRunList(v, r)
	case *run.Run:
		payload, err = m.toRun(v, r)
	case run.Phase:
		payload, err = m.toPhase(v, r)
	case *state.VersionList:
		payload = m.toList(v)
	case *state.Version:
		payload = m.toVersion(v)
	case state.OutputList:
		payload = m.toOutputList(v)
	case *state.Output:
		payload = m.toOutput(v)
	default:
		jsonapi.Error(w, fmt.Errorf("cannot marshal unknown type: %T", v))
		return
	}
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	jsonapi.WriteResponse(w, r, payload, opts...)
}
