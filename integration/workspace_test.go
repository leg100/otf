package integration

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/workspace"
)

func newWorkspaceService(t *testing.T, db otf.DB, repoService repo.Service) workspace.Service {
	return workspace.NewService(workspace.Options{
		Logger:      logr.Discard(),
		DB:          db,
		RepoService: repoService,
		Broker:      testBroker(t, db),
	})
}
