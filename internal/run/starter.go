package run

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
)

const (
	PlanOnlyOperation     Operation = "plan-only"
	PlanAndApplyOperation Operation = "plan-and-apply"
	DestroyAllOperation   Operation = "destroy-all"
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

	// Run operation specifies the terraform execution mode.
	Operation string
)

func (rs *starter) startRun(ctx context.Context, workspaceID string, op Operation) (*Run, error) {
	var (
		planOnly bool
		destroy  bool
	)

	switch op {
	case PlanOnlyOperation:
		planOnly = true
	case PlanAndApplyOperation:
		planOnly = false
	case DestroyAllOperation:
		destroy = true
	default:
		return nil, fmt.Errorf("invalid run operation: %s", op)
	}

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
		cv, err = rs.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{})
		if err != nil {
			return nil, err
		}
		if err := rs.UploadConfig(ctx, cv.ID, tarball); err != nil {
			return nil, err
		}
	} else {
		cv, err = rs.GetLatestConfigurationVersion(ctx, ws.ID)
		if err != nil {
			if errors.Is(err, internal.ErrResourceNotFound) {
				return nil, fmt.Errorf("missing configuration: you need to either start a run via terraform, or connect a repository")
			}
			return nil, err
		}
	}

	return rs.CreateRun(ctx, workspaceID, RunCreateOptions{
		ConfigurationVersionID: internal.String(cv.ID),
		IsDestroy:              &destroy,
		PlanOnly:               &planOnly,
	})
}
