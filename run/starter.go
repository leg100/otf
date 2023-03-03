package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/vcsprovider"
)

// RunStarter starts a run triggered via the UI (whereas the terraform CLI takes
// care of calling all the API endpoints to start a run itself).
type RunStarter struct {
	otf.ConfigurationVersionService
	*vcsprovider.Service
	otf.WorkspaceService

	service
}

func (rs *RunStarter) StartRun(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*otf.Run, error) {
	ws, err := rs.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv otf.ConfigurationVersion
	if ws.Repo != nil {
		client, err := rs.GetVCSClient(ctx, ws.Repo.ProviderID)
		if err != nil {
			return nil, err
		}
		tarball, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
			Identifier: ws.Repo.Identifier,
			Ref:        ws.Repo.Branch,
		})
		if err != nil {
			return nil, fmt.Errorf("retrieving repository tarball: %w", err)
		}
		cv, err = rs.CreateConfigurationVersion(ctx, ws.ID, opts)
		if err != nil {
			return nil, err
		}
		if err := rs.UploadConfig(ctx, cv.ID, tarball); err != nil {
			return nil, err
		}
	} else {
		latest, err := rs.GetLatestConfigurationVersion(ctx, ws.ID)
		if err != nil {
			if errors.Is(err, otf.ErrResourceNotFound) {
				return nil, fmt.Errorf("missing configuration: you need to either start a run via terraform, or connect a repository")
			}
			return nil, err
		}
		cv, err = rs.CloneConfigurationVersion(ctx, latest.ID, opts)
		if err != nil {
			return nil, err
		}
	}

	return rs.create(ctx, workspaceID, otf.RunCreateOptions{
		ConfigurationVersionID: otf.String(cv.ID),
	})
}
