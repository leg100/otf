package configversion

import "context"

type FakeService struct {
	cv *ConfigurationVersion
}

func (f *FakeService) GetConfigurationVersion(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *FakeService) GetLatestConfigurationVersion(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *FakeService) CreateConfigurationVersion(context.Context, string, ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	return &ConfigurationVersion{ID: "created"}, nil
}

func (f *FakeService) UploadConfig(context.Context, string, []byte) error {
	return nil
}
