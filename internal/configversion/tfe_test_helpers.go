package configversion

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

type fakeConfigService struct {
	*Service
}

func (f *fakeConfigService) UploadConfig(context.Context, resource.TfeID, []byte) error {
	return nil
}
