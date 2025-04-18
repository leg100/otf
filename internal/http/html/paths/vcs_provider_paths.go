// Code generated by "go generate"; DO NOT EDIT.

package paths

import "fmt"

func VCSProviders(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/vcs-providers", organization)
}

func CreateVCSProvider(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/vcs-providers/create", organization)
}

func NewVCSProvider(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/vcs-providers/new", organization)
}

func VCSProvider(vcsProvider fmt.Stringer) string {
	return fmt.Sprintf("/app/vcs-providers/%s", vcsProvider)
}

func EditVCSProvider(vcsProvider fmt.Stringer) string {
	return fmt.Sprintf("/app/vcs-providers/%s/edit", vcsProvider)
}

func UpdateVCSProvider(vcsProvider fmt.Stringer) string {
	return fmt.Sprintf("/app/vcs-providers/%s/update", vcsProvider)
}

func DeleteVCSProvider(vcsProvider fmt.Stringer) string {
	return fmt.Sprintf("/app/vcs-providers/%s/delete", vcsProvider)
}

func NewGithubAppVCSProvider(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/vcs-providers/new-github-app", organization)
}
