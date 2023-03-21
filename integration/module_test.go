package integration

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/vcsprovider"
)

func newModuleService(t *testing.T, db otf.DB, repoService repo.Service, vcsService vcsprovider.Service) module.Service {
	return module.NewService(module.Options{
		Logger:             logr.Discard(),
		DB:                 db,
		RepoService:        repoService,
		VCSProviderService: vcsService,
	})
}
