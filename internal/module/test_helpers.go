package module

import (
	"context"

	"github.com/leg100/otf/internal/vcs"
)

type fakeModulesCloudClient struct {
	repos []vcs.Repo

	vcs.Client
}

func (f *fakeModulesCloudClient) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]vcs.Repo, error) {
	return f.repos, nil
}
