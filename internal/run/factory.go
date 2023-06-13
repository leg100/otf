package run

import (
	"context"

	"github.com/leg100/otf/internal/configversion"
)

// factory constructs runs
type factory struct {
	ConfigurationVersionService
	WorkspaceService
}

// NewRun constructs a new run using the provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	ws, err := f.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv *configversion.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = f.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
	} else {
		cv, err = f.GetLatestConfigurationVersion(ctx, workspaceID)
	}
	if err != nil {
		return nil, err
	}

	return newRun(cv, ws, opts), nil
}

func (f *factory) createConfigVersionFromVCS(ws *Workspace) (*ConfigurationVersion, error) {
	if ws.Connection == nil {
		return nil, workspace.ErrNoVCSConnection
	}
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

