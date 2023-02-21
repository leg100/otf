package organization

import "github.com/leg100/otf/http/jsonapi"

func newFromJSONAPI(from jsonapi.Organization) *Organization {
	return &Organization{
		id:              from.ExternalID,
		createdAt:       from.CreatedAt,
		name:            from.Name,
		sessionRemember: from.SessionRemember,
		sessionTimeout:  from.SessionTimeout,
	}
}

// ToJSONAPI assembles a JSONAPI DTO
func toJSONAPI(org *Organization) *jsonapi.Organization {
	return &jsonapi.Organization{
		Name:            org.name,
		CreatedAt:       org.createdAt,
		ExternalID:      org.id,
		Permissions:     &jsonapi.DefaultOrganizationPermissions,
		SessionRemember: org.sessionRemember,
		SessionTimeout:  org.sessionTimeout,
	}
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *OrganizationList) listToJSONAPI(from *OrganizationList) *jsonapi.OrganizationList {
	to := &jsonapi.OrganizationList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		to.Items = append(to.Items, toJSONAPI(item))
	}
	return to
}
