package configversion

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

type FakeService struct {
	cv *ConfigurationVersion
}

func (f *FakeService) Get(context.Context, resource.TfeID) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *FakeService) GetLatest(context.Context, resource.TfeID) (*ConfigurationVersion, error) {
	return f.cv, nil
}

func (f *FakeService) Create(context.Context, resource.TfeID, CreateOptions) (*ConfigurationVersion, error) {
	return &ConfigurationVersion{}, nil
}

func (f *FakeService) UploadConfig(context.Context, resource.TfeID, []byte) error {
	return nil
}
