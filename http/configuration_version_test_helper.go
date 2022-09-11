package http

import (
	"context"

	"github.com/leg100/otf"
)

type fakeConfigurationVersionApp struct {
	otf.Application
}

func (f *fakeConfigurationVersionApp) UploadConfig(context.Context, string, []byte) error {
	return nil
}
