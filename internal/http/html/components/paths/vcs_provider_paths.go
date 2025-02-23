// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func VCSProviders(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/vcs-providers", organization))
}

func CreateVCSProvider(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/vcs-providers/create", organization))
}

func NewVCSProvider(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/vcs-providers/new", organization))
}

func VCSProvider(vcsProvider string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/vcs-providers/%s", vcsProvider))
}

func EditVCSProvider(vcsProvider string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/vcs-providers/%s/edit", vcsProvider))
}

func UpdateVCSProvider(vcsProvider string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/vcs-providers/%s/update", vcsProvider))
}

func DeleteVCSProvider(vcsProvider string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/vcs-providers/%s/delete", vcsProvider))
}

func NewGithubAppVCSProvider(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/vcs-providers/new-github-app", organization))
}
