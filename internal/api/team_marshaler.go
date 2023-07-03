package api

import (
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
)

func (m *jsonapiMarshaler) toTeam(from *auth.Team, r *http.Request) (*types.Team, []jsonapi.MarshalOption, error) {
	to := &types.Team{
		ID:         from.ID,
		Name:       from.Name,
		SSOTeamID:  from.SSOTeamID,
		Visibility: from.Visibility,
		OrganizationAccess: &types.OrganizationAccess{
			ManageWorkspaces:      from.Access.ManageWorkspaces,
			ManageVCSSettings:     from.Access.ManageVCS,
			ManageModules:         from.Access.ManageModules,
			ManageProviders:       from.Access.ManageProviders,
			ManagePolicies:        from.Access.ManagePolicies,
			ManagePolicyOverrides: from.Access.ManagePolicyOverrides,
		},
		// Hardcode these values until proper support is added
		Permissions: &types.TeamPermissions{
			CanDestroy:          true,
			CanUpdateMembership: true,
		},
	}

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
					to.Users = append(to.Users, m.toUser(user))
				}
				opts = append(opts, jsonapi.MarshalInclude(include...))
			}
		}
	}
	return to, opts, nil
}

func (m *jsonapiMarshaler) toTeamList(from []*auth.Team, r *http.Request) (to []*types.Team, opts []jsonapi.MarshalOption, err error) {
	var listOptions resource.PageOptions
	if err := decode.Query(&listOptions, r.URL.Query()); err != nil {
		return nil, nil, err
	}
	from, pagination := resource.Paginate(from, listOptions)
	meta := jsonapi.MarshalMeta(map[string]*types.Pagination{
		"meta": (*types.Pagination)(pagination),
	})
	opts = append(opts, jsonapi.MarshalOption(meta))

	to = make([]*types.Team, len(from))
	for i, fromTeam := range from {
		to[i], _, err = m.toTeam(fromTeam, r)
		if err != nil {
			return nil, nil, err
		}
	}

	return to, opts, nil
}
