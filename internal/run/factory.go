package run

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/workspace"
)

// factory constructs runs
type factory struct {
	ConfigurationVersionService
	WorkspaceService
	VCSProviderService
}

// NewRun constructs a new run using the provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	ws, err := f.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// There are two possibilities for the ConfigurationVersionID value:
	// (a) non-nil, in which case it is deemed to be a configuration version id
	// and an existing config version with that ID is retrieved.
	// (b) nil, in which it depends whether the workspace is connected to a vcs
	// repo:
	// 	(i) not connected: the latest config version is retrieved.
	// 	(ii) connected: same behaviour as (a): vcs repo contents are retrieved.
	var cv *configversion.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = f.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
	} else if ws.Connection == nil {
		cv, err = f.GetLatestConfigurationVersion(ctx, workspaceID)
	} else {
		cv, err = f.createConfigVersionFromVCS(ctx, ws)
	}
	if err != nil {
		return nil, err
	}

	return newRun(cv, ws, opts), nil
}

// createConfigVersionFromVCS creates a config version from the vcs repo
// connected to the workspace using the contents of the vcs repo.
func (f *factory) createConfigVersionFromVCS(ctx context.Context, ws *workspace.Workspace) (*configversion.ConfigurationVersion, error) {
	client, err := f.GetVCSClient(ctx, ws.Connection.VCSProviderID)
	if err != nil {
		return nil, err
	}
	// use workspace branch if set
	var ref *string
	if ws.Connection.Branch != "" {
		ref = &ws.Connection.Branch
	}
	tarball, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Repo: ws.Connection.Repo,
		Ref:  ref,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving repository tarball: %w", err)
	}
	cv, err := f.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{})
	if err != nil {
		return nil, err
	}
	if err := f.UploadConfig(ctx, cv.ID, tarball); err != nil {
		return nil, err
	}
	return cv, err
}
