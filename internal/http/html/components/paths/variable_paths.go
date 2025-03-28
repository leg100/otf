// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func Variables(workspace fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/variables", workspace))
}

func CreateVariable(workspace fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/variables/create", workspace))
}

func NewVariable(workspace fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/workspaces/%s/variables/new", workspace))
}

func Variable(variable fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variables/%s", variable))
}

func EditVariable(variable fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variables/%s/edit", variable))
}

func UpdateVariable(variable fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variables/%s/update", variable))
}

func DeleteVariable(variable fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/variables/%s/delete", variable))
}
