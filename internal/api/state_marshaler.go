package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/state"
)

func (m *jsonapiMarshaler) toState(from *state.Version, r *http.Request) (*types.StateVersion, []jsonapi.MarshalOption) {
	to := &types.StateVersion{
		ID:          from.ID,
		CreatedAt:   from.CreatedAt,
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", from.ID),
		Serial:      from.Serial,
	}
	for _, out := range from.Outputs {
		to.Outputs = append(to.Outputs, &types.StateVersionOutput{ID: out.ID})
	}

	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/state-versions#outputs
	var opts []jsonapi.MarshalOption
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "outputs":
				var include []any
				for _, out := range from.Outputs {
					include = append(include, m.toOutput(out))
				}
				opts = append(opts, jsonapi.MarshalInclude(include...))
			}
		}
	}
	return to, opts
}

func (*jsonapiMarshaler) toOutput(from *state.Output) *types.StateVersionOutput {
	to := &types.StateVersionOutput{
		ID:        from.ID,
		Name:      from.Name,
		Sensitive: from.Sensitive,
		Type:      from.Type,
		Value:     from.Value,
	}
	if to.Sensitive {
		to.Value = nil
	}
	return to
}
