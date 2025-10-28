package run

import (
	"context"
	"time"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

type (
	fakeFactoryOrganizationService struct {
		org *organization.Organization
	}
	fakeFactoryWorkspaceService struct {
		ws *workspace.Workspace
	}
	fakeFactoryConfigurationVersionService struct {
		cv *configversion.ConfigurationVersion
	}
	fakeFactoryVCSProviderService struct{}
	fakeFactoryCloudClient        struct {
		vcs.Client
	}
	fakeReleasesService struct {
		latestVersion string
	}
)

func newTestFactory(org *organization.Organization, ws *workspace.Workspace, cv *configversion.ConfigurationVersion) *factory {
	return &factory{
		organizations: &fakeFactoryOrganizationService{org: org},
		workspaces:    &fakeFactoryWorkspaceService{ws: ws},
		configs:       &fakeFactoryConfigurationVersionService{cv: cv},
		vcs:           &fakeFactoryVCSProviderService{},
	}
}

func (f *fakeFactoryOrganizationService) Get(context.Context, organization.Name) (*organization.Organization, error) {
	return f.org, nil
}

func (f *fakeFactoryWorkspaceService) Get(context.Context, resource.TfeID) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeFactoryConfigurationVersionService) Get(context.Context, resource.TfeID) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryConfigurationVersionService) GetLatest(context.Context, resource.TfeID) (*configversion.ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeFactoryConfigurationVersionService) Create(ctx context.Context, workspaceID resource.TfeID, opts configversion.CreateOptions) (*configversion.ConfigurationVersion, error) {
	return &configversion.ConfigurationVersion{ID: workspaceID}, nil
}

func (f *fakeFactoryConfigurationVersionService) UploadConfig(context.Context, resource.TfeID, []byte) error {
	return nil
}

func (f *fakeFactoryConfigurationVersionService) GetSourceIcon(source source.Source) templ.Component {
	return templ.Raw("")
}

func (f *fakeFactoryVCSProviderService) Get(context.Context, resource.TfeID) (*vcs.Provider, error) {
	return &vcs.Provider{
		Client: &fakeFactoryCloudClient{},
	}, nil
}

func (f *fakeFactoryCloudClient) GetRepoTarball(context.Context, vcs.GetRepoTarballOptions) ([]byte, string, error) {
	return nil, "", nil
}

func (f *fakeFactoryCloudClient) GetDefaultBranch(context.Context, string) (string, error) {
	return "", nil
}

func (f *fakeFactoryCloudClient) GetCommit(context.Context, vcs.Repo, string) (vcs.Commit, error) {
	return vcs.Commit{}, nil
}

func (f *fakeReleasesService) GetLatest(context.Context, *engine.Engine) (string, time.Time, error) {
	return f.latestVersion, time.Time{}, nil
}
