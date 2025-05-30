// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func Organizations() templ.SafeURL {
	return templ.URL("/app/organizations")
}

func CreateOrganization() templ.SafeURL {
	return templ.URL("/app/organizations/create")
}

func NewOrganization() templ.SafeURL {
	return templ.URL("/app/organizations/new")
}

func Organization(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s", organization))
}

func EditOrganization(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/edit", organization))
}

func UpdateOrganization(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/update", organization))
}

func DeleteOrganization(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/delete", organization))
}
