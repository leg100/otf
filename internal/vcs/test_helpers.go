package vcs

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
)

func NewTestRepo() Repo {
	return Repo{Owner: "repo-owner-" + internal.GenerateRandomString(4), Name: "repo-" + internal.GenerateRandomString(4)}
}

func NewTestModuleRepo(provider, name string) Repo {
	return Repo{Owner: uuid.NewString(), Name: fmt.Sprintf("terraform-%s-%s", provider, name)}
}
