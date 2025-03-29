package vcsprovider

import "github.com/leg100/otf/internal/resource"

func NewTestVCSProvider() *VCSProvider {
	return &VCSProvider{
		ID: resource.NewTfeID("vcs"),
	}
}
