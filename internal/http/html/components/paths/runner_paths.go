// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func Runners(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/runners", organization))
}

func CreateRunner(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/runners/create", organization))
}

func NewRunner(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/runners/new", organization))
}

func Runner(runner fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/runners/%s", runner))
}

func EditRunner(runner fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/runners/%s/edit", runner))
}

func UpdateRunner(runner fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/runners/%s/update", runner))
}

func DeleteRunner(runner fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/runners/%s/delete", runner))
}

func WatchRunners(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/runners/watch", organization))
}
