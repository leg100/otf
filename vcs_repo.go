package otf

import "strings"

type VCSRepo struct {
	// identifier is <repo_owner>/<repo_name>
	Identifier string
	// httpURL is the web url for the repo
	HttpURL string
	// VCSRepo belongs to a workspace
	WorkspaceID string
	// VCSRepo has a VCSProvider
	ProviderID string
	// Branch is repo's default mainline branch
	Branch string
}

func (r VCSRepo) String() string { return r.Identifier }
func (r VCSRepo) Owner() string  { return strings.Split(r.Identifier, "/")[0] }
func (r VCSRepo) Repo() string   { return strings.Split(r.Identifier, "/")[1] }
