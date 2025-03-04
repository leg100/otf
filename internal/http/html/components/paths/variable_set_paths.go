// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func VariableSets(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/variable-sets", organization))
}

func CreateVariableSet(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/variable-sets/create", organization))
}

func NewVariableSet(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/variable-sets/new", organization))
}

func VariableSet(variableSet string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s", variableSet))
}

func EditVariableSet(variableSet string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/edit", variableSet))
}

func UpdateVariableSet(variableSet string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/update", variableSet))
}

func DeleteVariableSet(variableSet string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/delete", variableSet))
}
