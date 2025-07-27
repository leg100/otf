package runner

import (
	"context"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
	"golang.org/x/sync/errgroup"
)

type RemoteOperationSpawner struct {
	Config OperationConfig
	Logger logr.Logger
	URL    string
	n      int
}

func (s *RemoteOperationSpawner) SpawnOperation(ctx context.Context, g *errgroup.Group, job *Job, jobToken []byte) error {
	// Construct an API client authenticating as the job.
	client, err := otfapi.NewClient(otfapi.Config{
		URL:           s.URL,
		Token:         string(jobToken),
		Logger:        s.Logger,
		RetryRequests: true,
	})
	if err != nil {
		return err
	}
	s.n++
	doOperation(ctx, g, operationOptions{
		logger:          s.Logger,
		OperationConfig: s.Config,
		job:             job,
		jobToken:        jobToken,
		runs:            &run.Client{Client: client},
		jobs:            &Client{Client: client},
		workspaces:      &workspace.Client{Client: client},
		variables:       &variable.Client{Client: client},
		state:           &state.Client{Client: client},
		configs:         &configversion.Client{Client: client},
		server:          client,
	})
	s.n--
	return nil
}

func (s *RemoteOperationSpawner) currentJobs() int {
	return s.n
}
