package otf

import (
	"context"
	"errors"
	"fmt"
)

// RunStarter starts a run triggered via the UI (whereas the terraform CLI takes
// care of calling all the API endpoints to start a run itself).
type RunStarter struct {
	Application
}

func (rs *RunStarter) StartRun(ctx context.Context, spec WorkspaceSpec, opts ConfigurationVersionCreateOptions) (*Run, error) {
	ws, err := rs.GetWorkspace(ctx, spec)
	if err != nil {
		return nil, err
	}

	var cv *ConfigurationVersion
	if ws.Repo() != nil {
		tarball, err := rs.GetRepoTarball(ctx, ws.Repo().ProviderID, GetRepoTarballOptions{
			Identifier: ws.Repo().Identifier,
			Ref:        ws.Repo().Branch,
		})
		if err != nil {
			return nil, fmt.Errorf("retrieving repository tarball: %w", err)
		}
		cv, err = rs.CreateConfigurationVersion(ctx, ws.ID(), opts)
		if err != nil {
			return nil, err
		}
		if err := rs.UploadConfig(ctx, cv.ID(), tarball); err != nil {
			return nil, err
		}
	} else {
		latest, err := rs.GetLatestConfigurationVersion(ctx, ws.ID())
		if err != nil {
			if errors.Is(err, ErrResourceNotFound) {
				return nil, fmt.Errorf("missing configuration: you need to either start a run via terraform, or connect a repository")
			}
			return nil, err
		}
		cv, err = rs.CloneConfigurationVersion(ctx, latest.ID(), opts)
		if err != nil {
			return nil, err
		}
	}

	return rs.CreateRun(ctx, spec, RunCreateOptions{
		ConfigurationVersionID: String(cv.ID()),
	})
}
