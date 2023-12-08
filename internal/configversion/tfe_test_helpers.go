package configversion

import "context"

type fakeConfigService struct {
	*Service
}

func (f *fakeConfigService) UploadConfig(context.Context, string, []byte) error {
	return nil
}
