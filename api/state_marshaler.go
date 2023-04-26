package api

import (
	"fmt"

	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/state"
)

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toVersion(from *state.Version) *jsonapi.StateVersion {
	to := &jsonapi.StateVersion{
		ID:          from.ID,
		CreatedAt:   from.CreatedAt,
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", from.ID),
		Serial:      from.Serial,
	}
	for _, out := range from.Outputs {
		to.Outputs = append(to.Outputs, m.toOutput(out))
	}
	return to
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toList(from *state.VersionList) *jsonapi.StateVersionList {
	jl := &jsonapi.StateVersionList{
		Pagination: jsonapi.NewPagination(from.Pagination),
	}
	for _, item := range from.Items {
		jl.Items = append(jl.Items, m.toVersion(item))
	}
	return jl
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (*jsonapiMarshaler) toOutput(from *state.Output) *jsonapi.StateVersionOutput {
	return &jsonapi.StateVersionOutput{
		ID:        from.ID,
		Name:      from.Name,
		Sensitive: from.Sensitive,
		Type:      from.Type,
		Value:     from.Value,
	}
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toOutputList(from state.OutputList) *jsonapi.StateVersionOutputList {
	var to jsonapi.StateVersionOutputList
	for _, v := range from {
		to.Items = append(to.Items, m.toOutput(v))
	}
	return &to
}
