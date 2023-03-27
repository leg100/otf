package state

import (
	"fmt"

	"github.com/leg100/otf/http/jsonapi"
)

type jsonapiMarshaler struct{}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toVersion(from *Version) *jsonapi.StateVersion {
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
func (m *jsonapiMarshaler) toList(from *VersionList) *jsonapi.StateVersionList {
	jl := &jsonapi.StateVersionList{
		Pagination: from.Pagination.ToJSONAPI(),
	}
	for _, item := range from.Items {
		jl.Items = append(jl.Items, m.toVersion(item))
	}
	return jl
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (*jsonapiMarshaler) toOutput(from *Output) *jsonapi.StateVersionOutput {
	return &jsonapi.StateVersionOutput{
		ID:        from.ID,
		Name:      from.Name,
		Sensitive: from.Sensitive,
		Type:      from.Type,
		Value:     from.Value,
	}
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toOutputList(from outputList) *jsonapi.StateVersionOutputList {
	var to jsonapi.StateVersionOutputList
	for _, v := range from {
		to.Items = append(to.Items, m.toOutput(v))
	}
	return &to
}
