package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/configversion"
)

// starter starts a run triggered via the UI (whereas the terraform CLI takes
// care of calling all the API endpoints to start a run itself).
type starter struct {
	ConfigurationVersionService
	WorkspaceService
	VCSProviderService
	RunService
}

func (rs *starter) startRun(ctx context.Context, workspaceID string, opts configversion.ConfigurationVersionCreateOptions) (*Run, error) {
	ws, err := rs.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv *configversion.ConfigurationVersion
	if ws.Connection != nil {
		client, err := rs.GetVCSClient(ctx, ws.Connection.VCSProviderID)
		if err != nil {
			return nil, err
		}
		// use workspace branch if set
		var ref *string
		if ws.Branch != "" {
			ref = &ws.Branch
		}
		tarball, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
			Repo: ws.Connection.Repo,
			Ref:  ref,
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

	return rs.CreateRun(ctx, workspaceID, RunCreateOptions{
		ConfigurationVersionID: otf.String(cv.ID),
	})
}
