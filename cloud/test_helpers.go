package cloud

import (
	"fmt"

	"github.com/google/uuid"
)

func NewTestRepo() Repo {
	identifier := uuid.NewString() + "/" + uuid.NewString()
	return Repo{
		Identifier: identifier,
		Branch:     "master",
	}
}

func NewTestModuleRepo(provider, name string) Repo {
	identifier := fmt.Sprintf("%s/terraform-%s-%s", uuid.New(), provider, name)
	return Repo{
		Identifier: identifier,
		Branch:     "master",
	}
}
