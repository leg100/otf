package configversion

import "context"

type FakeService struct {
	cv *ConfigurationVersion
}

func (f *FakeService) Get(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *FakeService) GetLatest(context.Context, string) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *FakeService) Create(context.Context, string, CreateOptions) (*ConfigurationVersion, error) {
	return &ConfigurationVersion{ID: "created"}, nil
}

func (f *FakeService) UploadConfig(context.Context, string, []byte) error {
	return nil
}
