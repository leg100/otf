package api

import (
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
)

func (m *jsonapiMarshaler) toTeam(from *auth.Team, r *http.Request) (*types.Team, []jsonapi.MarshalOption, error) {
	// Support including related resources:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/teams#available-related-resources
	var opts []jsonapi.MarshalOption
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "users":
				var include []any
				users, err := m.ListTeamMembers(r.Context(), from.ID)
				if err != nil {
					return nil, nil, err
				}
				for _, user := range users {
					include = append(include, m.toUser(user))
				}
				opts = append(opts, jsonapi.MarshalInclude(include...))
			}
		}
	}
	return &types.Team{
		ID:   from.ID,
		Name: from.Name,
	}, opts, nil
}
