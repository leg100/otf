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

	// There are three possibilities for the ConfigurationVersionID value:
	// (a) equals PullVCSMagicString in which case a config version is first
	// created from the contents of the vcs repo connected to the workspace
	// (b) non-nil, in which case it is deemed to be a configuration version id
	// and an existing config version with that ID is retrieved.
	// (c) nil, in which the latest config version is retrieved.
	var cv *configversion.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		if *opts.ConfigurationVersionID == PullVCSMagicString {
			cv, err = f.createConfigVersionFromVCS(ctx, ws)
		} else {
			cv, err = f.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
		}
	} else {
		cv, err = f.GetLatestConfigurationVersion(ctx, workspaceID)
	}
	if err != nil {
		return nil, err
	}

	return newRun(cv, ws, opts), nil
}

// createConfigVersionFromVCS creates a config version from the vcs repo
// connected to the workspace using the contents of the vcs repo.
func (f *factory) createConfigVersionFromVCS(ctx context.Context, ws *workspace.Workspace) (*configversion.ConfigurationVersion, error) {
	if ws.Connection == nil {
		return nil, workspace.ErrNoVCSConnection
	}
	client, err := f.GetVCSClient(ctx, ws.Connection.VCSProviderID)
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
	cv, err := f.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{})
	if err != nil {
		return nil, err
	}
	if err := f.UploadConfig(ctx, cv.ID, tarball); err != nil {
		return nil, err
	}
	return cv, err
}
