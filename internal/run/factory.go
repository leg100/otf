package run

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// factory constructs runs
	factory struct {
		organizations factoryOrganizationClient
		workspaces    factoryWorkspaceClient
		configs       factoryConfigClient
		vcs           factoryVCSClient
		releases      factoryReleasesClient
	}

	factoryOrganizationClient interface {
		GetOrganization(ctx context.Context, name string) (*organization.Organization, error)
	}

	factoryWorkspaceClient interface {
		Get(ctx context.Context, workspaceID string) (*workspace.Workspace, error)
	}

	factoryConfigClient interface {
		CreateConfigurationVersion(ctx context.Context, workspaceID string, opts configversion.ConfigurationVersionCreateOptions) (*configversion.ConfigurationVersion, error)
		GetConfigurationVersion(ctx context.Context, id string) (*configversion.ConfigurationVersion, error)
		GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (*configversion.ConfigurationVersion, error)
		UploadConfig(ctx context.Context, id string, config []byte) error
	}

	factoryVCSClient interface {
		GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error)
	}

	factoryReleasesClient interface {
		GetLatest(ctx context.Context) (string, time.Time, error)
	}
)

// NewRun constructs a new run using the provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts CreateOptions) (*Run, error) {
	ws, err := f.workspaces.Get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	org, err := f.organizations.GetOrganization(ctx, ws.Organization)
	if err != nil {
		return nil, err
	}
	if ws.TerraformVersion == releases.LatestVersionString {
		ws.TerraformVersion, _, err = f.releases.GetLatest(ctx)
		if err != nil {
			return nil, err
		}
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
		cv, err = f.configs.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
	} else if ws.Connection == nil {
		cv, err = f.configs.GetLatestConfigurationVersion(ctx, workspaceID)
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
	client, err := f.vcs.GetVCSClient(ctx, ws.Connection.VCSProviderID)
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
	cv, err := f.configs.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{
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
	if err := f.configs.UploadConfig(ctx, cv.ID, tarball); err != nil {
		return nil, err
	}
	return cv, err
}
