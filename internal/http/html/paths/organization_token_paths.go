// Code generated by "go generate"; DO NOT EDIT.

package paths

import "fmt"

func OrganizationToken(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/tokens/show", organization)
}

func CreateOrganizationToken(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/tokens/create", organization)
}

func DeleteOrganizationToken(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/tokens/delete", organization)
}
