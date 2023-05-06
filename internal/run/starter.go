package run

import (
	"context"
	"errors"
	"fmt"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/configversion"
)

const (
	planOnly     runStrategy = "plan-only"
	planAndApply runStrategy = "plan-and-apply"
	destroyAll   runStrategy = "destroy-all"
)

type (
	// starter starts a run triggered via the UI (whereas the terraform CLI takes
	// care of calling all the API endpoints to start a run itself).
	starter struct {
		ConfigurationVersionService
		WorkspaceService
		VCSProviderService
		RunService
	}

	runStrategy string
)

func (rs *starter) startRun(ctx context.Context, workspaceID string, strategy runStrategy) (*Run, error) {
	var (
		speculative bool
		destroy     bool
	)

	switch strategy {
	case planOnly:
		speculative = true
	case planAndApply:
		speculative = false
	case destroyAll:
		destroy = true
	default:
		return nil, fmt.Errorf("invalid strategy: %s", strategy)
	}

	ws, err := rs.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv *configversion.ConfigurationVersion
	configOptions := configversion.ConfigurationVersionCreateOptions{
		Speculative: &speculative,
	}
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
		cv, err = rs.CreateConfigurationVersion(ctx, ws.ID, configOptions)
		if err != nil {
			return nil, err
		}
		if err := rs.UploadConfig(ctx, cv.ID, tarball); err != nil {
			return nil, err
		}
	} else {
		latest, err := rs.GetLatestConfigurationVersion(ctx, ws.ID)
		if err != nil {
			if errors.Is(err, internal.ErrResourceNotFound) {
				return nil, fmt.Errorf("missing configuration: you need to either start a run via terraform, or connect a repository")
			}
			return nil, err
		}
		cv, err = rs.CloneConfigurationVersion(ctx, latest.ID, configOptions)
		if err != nil {
			return nil, err
		}
	}

	return rs.CreateRun(ctx, workspaceID, RunCreateOptions{
		ConfigurationVersionID: internal.String(cv.ID),
		IsDestroy:              &destroy,
	})
}
