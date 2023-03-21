package integration

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/vcsprovider"
)

func newVCSProviderService(t *testing.T, db otf.DB, cloudService cloud.Service) vcsprovider.Service {
	return vcsprovider.NewService(vcsprovider.Options{
		DB:           db,
		Logger:       logr.Discard(),
		CloudService: cloudService,
	})
}
