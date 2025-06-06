// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func OrganizationToken(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/tokens/show", organization))
}

func CreateOrganizationToken(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/tokens/create", organization))
}

func DeleteOrganizationToken(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/tokens/delete", organization))
}
