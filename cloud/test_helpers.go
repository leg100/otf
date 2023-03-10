package cloud

import (
	"fmt"

	"github.com/google/uuid"
)

func NewTestRepo() string {
	return uuid.NewString() + "/" + uuid.NewString()
}

func NewTestModuleRepo(provider, name string) string {
	return fmt.Sprintf("%s/terraform-%s-%s", uuid.New(), provider, name)
}
