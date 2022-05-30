package otf

import "context"

type fakeRunFactoryWorkspaceService struct {
	ws *Workspace
	WorkspaceService
}

func (f *fakeRunFactoryWorkspaceService) Get(context.Context, WorkspaceSpec) (*Workspace, error) {
	return f.ws, nil
}

type fakeRunFactoryConfigurationVersionService struct {
	cv *ConfigurationVersion
	ConfigurationVersionService
}

func (f *fakeRunFactoryConfigurationVersionService) Get(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *fakeRunFactoryConfigurationVersionService) GetLatest(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}
