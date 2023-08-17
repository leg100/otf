package configversion

import "context"

type fakeConfigService struct {
	ConfigurationVersionService
}

func (f *fakeConfigService) UploadConfig(context.Context, string, []byte) error {
	return nil
}
