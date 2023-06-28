package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

const mediaType = "application/vnd.api+json"

type (
	marshaler interface {
		toRun(run *run.Run, r *http.Request) (*types.Run, []jsonapi.MarshalOption, error)
		toRunList(from *run.RunList, r *http.Request) ([]*types.Run, []jsonapi.MarshalOption, error)
		toPhase(from run.Phase, r *http.Request) (any, error)
		toPlan(plan run.Phase, r *http.Request) (*types.Plan, error)
		writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter))
	}

	jsonapiMarshaler struct {
		run.RunService
		organization.OrganizationService
		state.StateService
		workspace.WorkspaceService
		auth.TeamService

		*runLogsURLGenerator
	}
)

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (m *jsonapiMarshaler) writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter)) {
	var (
		payload     any
		marshalOpts []jsonapi.MarshalOption
		err         error
	)

	switch v := v.(type) {
	case *organization.OrganizationList:
		payload, marshalOpts = m.toOrganizationList(v)
	case *organization.Organization:
		payload = m.toOrganization(v)
	case organization.Entitlements:
		payload = (*types.Entitlements)(&v)
	case *workspace.WorkspaceList:
		payload, marshalOpts, err = m.toWorkspaceList(v, r)
	case *workspace.Workspace:
		payload, marshalOpts, err = m.toWorkspace(v, r)
	case *run.RunList:
		payload, marshalOpts, err = m.toRunList(v, r)
	case *run.Run:
		payload, marshalOpts, err = m.toRun(v, r)
	case *configversion.ConfigurationVersionList:
		payload, marshalOpts, err = m.toConfigurationVersionList(v)
	case *configversion.ConfigurationVersion:
		payload, err = m.toConfigurationVersion(v)
	case []*variable.Variable:
		payload = m.toVariableList(v)
	case *variable.Variable:
		payload = m.toVariable(v)
	case []*notifications.Config:
		payload = m.toNotificationList(v)
	case *notifications.Config:
		payload = m.toNotificationConfig(v)
	case run.Phase:
		payload, err = m.toPhase(v, r)
	case *state.VersionList:
		payload, marshalOpts = m.toStateList(v, r)
	case *state.Version:
		payload, marshalOpts = m.toState(v, r)
	case state.OutputList:
		payload = m.toOutputList(v)
	case *state.Output:
		payload = m.toOutput(v)
	case *auth.User:
		payload = m.toUser(v)
	case *auth.Team:
		payload, marshalOpts, err = m.toTeam(v, r)
	case *workspace.TagList:
		payload, marshalOpts = m.toTags(v)
	default:
		Error(w, fmt.Errorf("cannot marshal unknown type: %T", v))
		return
	}
	if err != nil {
		Error(w, err)
		return
	}

	b, err := jsonapi.Marshal(payload, marshalOpts...)
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	for _, o := range opts {
		o(w)
	}
	w.Write(b)
}

// WithCode is a helper func for writing an HTTP status code to a response
// stream. For use with writeResponse.
func withCode(code int) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.WriteHeader(code)
	}
}

func unmarshal(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return jsonapi.Unmarshal(b, v)
}
