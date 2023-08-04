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
	OrganizationService
	WorkspaceService
	ConfigurationVersionService
	VCSProviderService
}

// NewRun constructs a new run using the provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts CreateOptions) (*Run, error) {
	ws, err := f.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	org, err := f.GetOrganization(ctx, ws.Organization)
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

	return newRun(ctx, org, cv, ws, opts), nil
}

// createConfigVersionFromVCS creates a config version from the vcs repo
// connected to the workspace using the contents of the vcs repo.
func (f *factory) createConfigVersionFromVCS(ctx context.Context, ws *workspace.Workspace) (*configversion.ConfigurationVersion, error) {
	client, err := f.GetVCSClient(ctx, ws.Connection.VCSProviderID)
	if err != nil {
		return nil, err
	}
	repo, err := client.GetRepository(ctx, ws.Connection.Repo)
	if err != nil {
		return nil, fmt.Errorf("retrieving repository info: %w", err)
	}
	branch := ws.Connection.Branch
	if branch == "" {
		branch = repo.DefaultBranch
	}
	tarball, ref, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Repo: ws.Connection.Repo,
		Ref:  &branch,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving repository tarball: %w", err)
	}
	commit, err := client.GetCommit(ctx, ws.Connection.Repo, ref)
	if err != nil {
		return nil, fmt.Errorf("retrieving commit information: %s: %w", ref, err)
	}
	cv, err := f.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{
		IngressAttributes: &configversion.IngressAttributes{
			Branch:          branch,
			CommitSHA:       commit.SHA,
			CommitURL:       commit.URL,
			Repo:            ws.Connection.Repo,
			IsPullRequest:   false,
			OnDefaultBranch: branch == repo.DefaultBranch,
			SenderUsername:  commit.Author.Username,
			SenderAvatarURL: commit.Author.AvatarURL,
			SenderHTMLURL:   commit.Author.ProfileURL,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := f.UploadConfig(ctx, cv.ID, tarball); err != nil {
		return nil, err
	}
	return cv, err
}
