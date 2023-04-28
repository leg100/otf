package api

import (
	"fmt"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/state"
)

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toVersion(from *state.Version) *StateVersion {
	to := &StateVersion{
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
func (m *jsonapiMarshaler) toList(from *state.VersionList) (to []*StateVersion, opts []jsonapi.MarshalOption) {
	opts = []jsonapi.MarshalOption{toMarshalOption(from.Pagination)}
	for _, item := range from.Items {
		to = append(to, m.toVersion(item))
	}
	return
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (*jsonapiMarshaler) toOutput(from *state.Output) *StateVersionOutput {
	return &StateVersionOutput{
		ID:        from.ID,
		Name:      from.Name,
		Sensitive: from.Sensitive,
		Type:      from.Type,
		Value:     from.Value,
	}
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (m *jsonapiMarshaler) toOutputList(from state.OutputList) *StateVersionOutputList {
	var to StateVersionOutputList
	for _, v := range from {
		to.Items = append(to.Items, m.toOutput(v))
	}
	return &to
}
