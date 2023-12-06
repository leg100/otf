package module

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

type fakeService struct {
	mod      *Module
	tarball  []byte
	vcsprovs []*vcsprovider.VCSProvider
	repos    []string
	hostname string

	Service
	internal.HostnameService

	vcsprovider.VCSProviderService
}

func (f *fakeService) PublishModule(context.Context, PublishOptions) (*Module, error) {
	return f.mod, nil
}

func (f *fakeService) GetModuleByID(context.Context, string) (*Module, error) {
	return f.mod, nil
}

func (f *fakeService) DeleteModule(context.Context, string) (*Module, error) {
	return f.mod, nil
}

func (f *fakeService) ListModules(context.Context, ListModulesOptions) ([]*Module, error) {
	return []*Module{f.mod}, nil
}

func (f *fakeService) GetVCSProvider(context.Context, string) (*vcsprovider.VCSProvider, error) {
	return f.vcsprovs[0], nil
}

func (f *fakeService) ListVCSProviders(context.Context, string) ([]*vcsprovider.VCSProvider, error) {
	return f.vcsprovs, nil
}

func (f *fakeService) GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error) {
	return &fakeModulesCloudClient{repos: f.repos}, nil
}

func (f *fakeService) GetModuleInfo(context.Context, string) (*TerraformModule, error) {
	return unmarshalTerraformModule(f.tarball)
}

func (f *fakeService) Hostname() string {
	return f.hostname
}

type fakeModulesCloudClient struct {
	repos []string

	vcs.Client
}

func (f *fakeModulesCloudClient) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]string, error) {
	return f.repos, nil
}
