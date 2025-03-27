// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/resource"
)

func VariableSetVariables(variableSet resource.ID) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/variable-set-variables", variableSet))
}

func CreateVariableSetVariable(variableSet resource.ID) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/variable-set-variables/create", variableSet))
}

func NewVariableSetVariable(variableSet resource.ID) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-sets/%s/variable-set-variables/new", variableSet))
}

func VariableSetVariable(variableSetVariable resource.ID) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-set-variables/%s", variableSetVariable))
}

func EditVariableSetVariable(variableSetVariable resource.ID) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-set-variables/%s/edit", variableSetVariable))
}

func UpdateVariableSetVariable(variableSetVariable resource.ID) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-set-variables/%s/update", variableSetVariable))
}

func DeleteVariableSetVariable(variableSetVariable resource.ID) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variable-set-variables/%s/delete", variableSetVariable))
}
