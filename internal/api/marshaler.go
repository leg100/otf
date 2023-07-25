package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
)

const mediaType = "application/vnd.api+json"

type (
	marshaler interface {
		toRun(run *run.Run, r *http.Request) (*types.Run, []jsonapi.MarshalOption, error)
		toPhase(from run.Phase, r *http.Request) (any, error)
		toPlan(plan run.Phase, r *http.Request) (*types.Plan, error)
		toConfigurationVersion(from *configversion.ConfigurationVersion) *types.ConfigurationVersion
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
	payload, marshalOpts, err := m.convert(r, v)
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

// convert v into a struct suitable for encoding into json:api, along with
// json:api marshaling options.
func (m *jsonapiMarshaler) convert(r *http.Request, v any) (any, []jsonapi.MarshalOption, error) {
	var opts []jsonapi.MarshalOption

	// v is either a:
	//
	// (a) a resource.Page[T], containing pagination metadata and a slice of
	// resources
	// (b) a slice of resources
	// (c) a resource

	// Get the value of v so we can check its structure
	dst := reflect.Indirect(reflect.ValueOf(v))

	// Check if v is a page struct with both pagination and items fields
	if dst.Kind() == reflect.Struct {
		if pagination := dst.FieldByName("Pagination"); pagination.IsValid() {
			items := dst.FieldByName("Items")
			if !items.IsValid() {
				return nil, nil, errors.New("page has a pagination field but no items field")
			}
			opts = []jsonapi.MarshalOption{
				jsonapi.MarshalMeta(map[string]*types.Pagination{
					"pagination": (*types.Pagination)(pagination.Addr().Elem().Interface().(*resource.Pagination)),
				}),
			}
			// assign page items to destination for conversion below
			dst = items
		}
	}
	if dst.Kind() == reflect.Slice {
		var payload []any
		for i := 0; i < dst.Len(); i++ {
			v, o, err := m.convert(r, dst.Index(i).Interface())
			if err != nil {
				return nil, nil, err
			}
			payload = append(payload, v)
			opts = append(opts, o...)
		}
		return payload, opts, nil
	}

	var (
		payload any
		err     error
	)
	switch v := v.(type) {
	case *organization.Organization:
		payload = m.toOrganization(v)
	case organization.Entitlements:
		payload = (*types.Entitlements)(&v)
	case *workspace.Workspace:
		payload, opts, err = m.toWorkspace(v, r)
	case *run.Run:
		payload, opts, err = m.toRun(v, r)
	case *configversion.ConfigurationVersion:
		payload = m.toConfigurationVersion(v)
	case *variable.Variable:
		payload = m.toVariable(v)
	case *notifications.Config:
		payload = m.toNotificationConfig(v)
	case run.Phase:
		payload, err = m.toPhase(v, r)
	case *state.Version:
		payload, opts = m.toState(v, r)
	case *state.Output:
		payload = m.toOutput(v)
	case *auth.User:
		payload = m.toUser(v)
	case *auth.Team:
		payload, opts, err = m.toTeam(v, r)
	case *workspace.Tag:
		payload = m.toTag(v)
	case *vcsprovider.VCSProvider:
		payload = m.toOAuthClient(v)
	default:
		return nil, nil, fmt.Errorf("cannot marshal unknown type: %T", v)
	}
	if err != nil {
		return nil, nil, err
	}
	return payload, opts, nil
}

// withCode is a helper func for writing an HTTP status code to a response
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
