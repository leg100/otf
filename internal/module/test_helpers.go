package module

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
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

func (f *fakeService) ListModules(context.Context, ListModulesOptions) ([]*Module, error) {
	return []*Module{f.mod}, nil
}

func (f *fakeService) Get(context.Context, resource.TfeID) (*vcsprovider.VCSProvider, error) {
	return f.vcsprovs[0], nil
}

func (f *fakeService) List(context.Context, resource.OrganizationName) ([]*vcsprovider.VCSProvider, error) {
	return f.vcsprovs, nil
}

func (f *fakeService) GetVCSClient(ctx context.Context, providerID resource.TfeID) (vcs.Client, error) {
	return &fakeModulesCloudClient{repos: f.repos}, nil
}

func (f *fakeService) GetModuleInfo(context.Context, resource.TfeID) (*TerraformModule, error) {
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
