package run

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// factory constructs runs
	factory struct {
		client serviceClient
	}
)

// NewRun constructs a new run using the provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID resource.ID, opts CreateOptions) (*Run, error) {
	ws, err := f.client.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	org, err := f.client.GetOrganization(ctx, ws.Organization)
	if err != nil {
		return nil, err
	}
	opts.costEstimationEnabled = org.CostEstimationEnabled

	// retrieve or create config: if a config version ID is specified then
	// retrieve that; otherwise if the workspace is connected then the latest
	// config is retrieved from the connected vcs repo, and if the workspace is
	// not connected then the latest existing config is used.
	var cv *configversion.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = f.client.GetConfigVersion(ctx, *opts.ConfigurationVersionID)
	} else if ws.Connection == nil {
		cv, err = f.client.GetLatestConfigVersion(ctx, workspaceID)
	} else {
		cv, err = f.createConfigVersionFromVCS(ctx, ws)
	}
	if err != nil {
		return nil, fmt.Errorf("fetching configuration: %w", err)
	}

	// If workspace tracks the latest version then fetch it from db.
	if ws.EngineVersion.Latest {
		opts.EngineVersion, _, err = f.client.GetLatest(ctx, ws.Engine)
		if err != nil {
			return nil, err
		}
	} else {
		opts.EngineVersion = ws.EngineVersion.String()
	}

	if creator, _ := user.UserFromContext(ctx); creator != nil {
		opts.CreatedBy = &creator.Username
	}

	return NewRun(ws, cv, opts)
}

// createConfigVersionFromVCS creates a config version from the vcs repo
// connected to the workspace using the contents of the vcs repo.
func (f *factory) createConfigVersionFromVCS(ctx context.Context, ws *workspace.Workspace) (*configversion.ConfigurationVersion, error) {
	client, err := f.client.GetVCSProvider(ctx, ws.Connection.VCSProviderID)
	if err != nil {
		return nil, err
	}
	defaultBranch, err := client.GetDefaultBranch(ctx, ws.Connection.Repo.String())
	if err != nil {
		return nil, fmt.Errorf("retrieving repository info: %w", err)
	}
	branch := ws.Connection.Branch
	if branch == "" {
		branch = defaultBranch
	}
	tarball, ref, err := client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
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
	cv, err := f.client.CreateConfigVersion(ctx, ws.ID, configversion.CreateOptions{
		IngressAttributes: &configversion.IngressAttributes{
			Branch:          branch,
			CommitSHA:       commit.SHA,
			CommitURL:       commit.URL,
			Repo:            ws.Connection.Repo,
			IsPullRequest:   false,
			OnDefaultBranch: branch == defaultBranch,
			SenderUsername:  commit.Author.Username,
			SenderAvatarURL: commit.Author.AvatarURL,
			SenderHTMLURL:   commit.Author.ProfileURL,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := f.client.UploadConfig(ctx, cv.ID, tarball); err != nil {
		return nil, err
	}
	return cv, err
}
