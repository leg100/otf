package cloud

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
)

func NewTestRepo() string {
	return "repo-owner-" + internal.GenerateRandomString(4) + "/" + "repo-" + internal.GenerateRandomString(4)
}

func NewTestModuleRepo(provider, name string) string {
	return fmt.Sprintf("%s/terraform-%s-%s", uuid.New(), provider, name)
}
