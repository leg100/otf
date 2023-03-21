package run

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/workspace"
)

type (
	// spawner spawns new runs in response to vcs events
	spawner struct {
		logr.Logger

		ConfigurationVersionService
		WorkspaceService
		VCSProviderService
		RunService
	}

	SpawnerOptions struct {
		logr.Logger

		otf.Subscriber

		ConfigurationVersionService
		WorkspaceService
		VCSProviderService
		RunService
	}
)

// StartSpawner starts the run spawner.
func StartSpawner(ctx context.Context, opts SpawnerOptions) error {
	sub, err := opts.Subscriber.Subscribe(ctx, "run-spawner")
	if err != nil {
		return err
	}

	s := &spawner{
		Logger:                      opts.Logger.WithValues("component", "spawner"),
		ConfigurationVersionService: opts.ConfigurationVersionService,
		WorkspaceService:            opts.WorkspaceService,
		VCSProviderService:          opts.VCSProviderService,
		RunService:                  opts.RunService,
	}

	for {
		select {
		case event := <-sub:
			// skip non-vcs events
			if event.Type != otf.EventVCS {
				continue
			}

			if err := s.handle(ctx, event.Payload); err != nil {
				s.Error(err, "handling vcs event")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *spawner) handle(ctx context.Context, event cloud.VCSEvent) error {
	var repoID uuid.UUID
	var isPullRequest bool
	var branch, defaultBranch string
	var sha string

	switch event := event.(type) {
	case cloud.VCSPushEvent:
		repoID = event.RepoID
		sha = event.CommitSHA
		branch = event.Branch
		defaultBranch = event.DefaultBranch
	case cloud.VCSPullEvent:
		// only spawn runs when opening a PR and pushing to a PR
		switch event.Action {
		case cloud.VCSPullEventOpened, cloud.VCSPullEventUpdated:
		default:
			return nil
		}
		repoID = event.RepoID
		sha = event.CommitSHA
		branch = event.Branch
		defaultBranch = event.DefaultBranch
		isPullRequest = true
	}

	h.Info("spawning run", "repo_id", repoID)

	workspaces, err := h.ListWorkspacesByRepoID(ctx, repoID)
	if err != nil {
		return err
	}
	if len(workspaces) == 0 {
		h.Info("no connected workspaces found")
		return nil
	}

	// we have 1+ workspaces connected to this repo but we only need to retrieve
	// the repo once, and to do so we'll use the VCS provider associated with
	// the first workspace (any would do).
	if workspaces[0].Connection == nil {
		return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID)
	}
	providerID := workspaces[0].Connection.VCSProviderID

	client, err := h.GetVCSClient(ctx, providerID)
	if err != nil {
		return err
	}
	tarball, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Repo: workspaces[0].Connection.Repo,
		Ref:  &sha,
	})
	if err != nil {
		return fmt.Errorf("retrieving repository tarball: %w", err)
	}

	// Determine which workspaces to spawn runs for. If its a PR then a
	// (speculative) run is
	// spawned for all workspaces. Otherwise each workspace's branch setting
	// is checked and if it set then it must match the event's branch. If it is
	// not set then the event's branch must match the repo's default branch. If
	// neither of these conditions are true then the workspace is skipped.
	filterFunc := func(unfiltered []*workspace.Workspace) (filtered []*workspace.Workspace) {
		for _, ws := range unfiltered {
			if ws.Branch != "" && ws.Branch == branch {
				filtered = append(filtered, ws)
			} else if branch == defaultBranch {
				filtered = append(filtered, ws)
			} else {
				continue
			}
		}
		return
	}
	if !isPullRequest {
		workspaces = filterFunc(workspaces)
	}

	// create a config version for each workspace and spawn run.
	for _, ws := range workspaces {
		if ws.Connection == nil {
			// Should never happen...
			return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID)
		}
		cv, err := h.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{
			Speculative: otf.Bool(isPullRequest),
			IngressAttributes: &configversion.IngressAttributes{
				// ID     string
				Branch: branch,
				// CloneURL          string
				// CommitMessage     string
				CommitSHA: sha,
				// CommitURL         string
				// CompareURL        string
				Repo:            ws.Connection.Repo,
				IsPullRequest:   isPullRequest,
				OnDefaultBranch: branch == defaultBranch,
			},
		})
		if err != nil {
			return err
		}
		if err := h.UploadConfig(ctx, cv.ID, tarball); err != nil {
			return err
		}
		_, err = h.CreateRun(ctx, ws.ID, RunCreateOptions{
			ConfigurationVersionID: otf.String(cv.ID),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
