// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

func VariableSets(organization resource.ID) string {
	return fmt.Sprintf("/app/organizations/%s/variable-sets", organization)
}

func CreateVariableSet(organization resource.ID) string {
	return fmt.Sprintf("/app/organizations/%s/variable-sets/create", organization)
}

func NewVariableSet(organization resource.ID) string {
	return fmt.Sprintf("/app/organizations/%s/variable-sets/new", organization)
}

func VariableSet(variableSet resource.ID) string {
	return fmt.Sprintf("/app/variable-sets/%s", variableSet)
}

func EditVariableSet(variableSet resource.ID) string {
	return fmt.Sprintf("/app/variable-sets/%s/edit", variableSet)
}

func UpdateVariableSet(variableSet resource.ID) string {
	return fmt.Sprintf("/app/variable-sets/%s/update", variableSet)
}

func DeleteVariableSet(variableSet resource.ID) string {
	return fmt.Sprintf("/app/variable-sets/%s/delete", variableSet)
}
