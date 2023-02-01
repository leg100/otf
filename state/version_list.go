package state

import "github.com/leg100/otf"

// versionList represents a list of state versions.
type versionList struct {
	*otf.Pagination
	Items []*version
}

// ToJSONAPI assembles a struct suitable for marshalling into json-api
func (l *versionList) ToJSONAPI() any {
	jl := &jsonapiVersionList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		jl.Items = append(jl.Items, item.ToJSONAPI().(*jsonapiVersion))
	}
	return jl
}
