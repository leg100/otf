package otf

import "strings"

type VCSRepo struct {
	// identifier is <repo_owner>/<repo_name>
	Identifier string
	// HTTPURL is the web url for the repo
	HTTPURL string
	// VCSRepo has a VCSProvider
	ProviderID string
	// Branch is repo's default mainline branch
	Branch string
}

func NewVCSRepo(cloud Cloud) *VCSRepo {
	return nil
}

func (r VCSRepo) String() string { return r.Identifier }
func (r VCSRepo) Owner() string  { return strings.Split(r.Identifier, "/")[0] }
func (r VCSRepo) Repo() string   { return strings.Split(r.Identifier, "/")[1] }
