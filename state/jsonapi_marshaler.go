package state

import (
	"fmt"

	"github.com/leg100/otf/http/jsonapi"
)

type jsonapiMarshaler struct{}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toVersion(form *version) *jsonapi.StateVersion {
	to := &jsonapi.StateVersion{
		ID:          form.ID,
		CreatedAt:   form.CreatedAt,
		DownloadURL: fmt.Sprintf("/api/v2/state-versions/%s/download", form.ID),
		Serial:      form.Serial,
	}
	for _, out := range form.Outputs {
		to.Outputs = append(to.Outputs, m.toOutput(out))
	}
	return to
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toList(from *versionList) *jsonapi.StateVersionList {
	jl := &jsonapi.StateVersionList{
		Pagination: from.Pagination.ToJSONAPI(),
	}
	for _, item := range from.Items {
		jl.Items = append(jl.Items, m.toVersion(item))
	}
	return jl
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (*jsonapiMarshaler) toOutput(from *output) *jsonapi.StateVersionOutput {
	return &jsonapi.StateVersionOutput{
		ID:        from.id,
		Name:      from.name,
		Sensitive: from.sensitive,
		Type:      from.typ,
		Value:     from.value,
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
