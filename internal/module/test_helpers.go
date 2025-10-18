package module

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
)

type fakeService struct {
	mod       *Module
	tarball   []byte
	vcsprov   *vcs.Provider
	vcsprovs  []*vcs.Provider
	repos     []string
	hostname  string
	providers []string

	Service
	internal.HostnameService
}

func (f *fakeService) PublishModule(context.Context, PublishOptions) (*Module, error) {
	return f.mod, nil
}

func (f *fakeService) GetModuleByID(context.Context, resource.TfeID) (*Module, error) {
	return f.mod, nil
}

func (f *fakeService) DeleteModule(context.Context, resource.TfeID) (*Module, error) {
	return f.mod, nil
}

func (f *fakeService) ListModules(context.Context, ListOptions) ([]*Module, error) {
	return []*Module{f.mod}, nil
}

func (f *fakeService) Get(context.Context, resource.TfeID) (*vcs.Provider, error) {
	return f.vcsprov, nil
}

func (f *fakeService) List(context.Context, organization.Name) ([]*vcs.Provider, error) {
	return f.vcsprovs, nil
}

func (f *fakeService) GetModuleInfo(context.Context, resource.TfeID) (*TerraformModule, error) {
	return UnmarshalTerraformModule(f.tarball)
}

func (f *fakeService) Hostname() string {
	return f.hostname
}

type fakeModulesCloudClient struct {
	repos []vcs.Repo

	vcs.Client
}

func (f *fakeModulesCloudClient) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]vcs.Repo, error) {
	return f.repos, nil
}
