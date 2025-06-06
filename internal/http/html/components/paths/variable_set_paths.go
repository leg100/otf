// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func VariableSets(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/variable-sets", organization))
}

func CreateVariableSet(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/variable-sets/create", organization))
}

func NewVariableSet(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/variable-sets/new", organization))
}

func VariableSet(variableSet fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s", variableSet))
}

func EditVariableSet(variableSet fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/edit", variableSet))
}

func UpdateVariableSet(variableSet fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/update", variableSet))
}

func DeleteVariableSet(variableSet fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/delete", variableSet))
}
