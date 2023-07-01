package api

import (
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/decode"
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
	var listOptions internal.ListOptions
	if err := decode.All(&listOptions, r); err != nil {
		return nil, nil, err
	}
	pagination := internal.NewPagination(listOptions, len(from))
	opts = []jsonapi.MarshalOption{toMarshalOption(pagination)}

	// trim the start
	start := listOptions.SanitizedPageSize() * (listOptions.SanitizedPageNumber() - 1)
	if start > len(from) {
		// paging is out-of-range: return empty list
		return to, opts, nil
	}
	from = from[start:]

	// trim the end
	end := listOptions.SanitizedPageSize() * listOptions.SanitizedPageNumber()
	if len(from) > end {
		from = from[:(end - 1)]
	}

	to = make([]*types.Team, len(from))
	for i, fromTeam := range from {
		to[i], _, err = m.toTeam(fromTeam, r)
		if err != nil {
			return nil, nil, err
		}
	}

	return to, opts, nil
}
