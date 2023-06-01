package run

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// Spawner spawns new runs in response to vcs events
	Spawner struct {
		logr.Logger

		ConfigurationVersionService
		WorkspaceService
		VCSProviderService
		RunService
		pubsub.Subscriber
	}
)

// Start the run spawner.
func (s *Spawner) Start(ctx context.Context) error {
	sub, err := s.Subscribe(ctx, "run-spawner")
	if err != nil {
		return err
	}

	for event := range sub {
		// skip non-vcs events
		if event.Type != pubsub.EventVCS {
			continue
		}
		if err := s.handle(ctx, event.Payload); err != nil {
			s.Error(err, "handling vcs event")
		}
	}
	return nil
}

func (s *Spawner) handle(ctx context.Context, event cloud.VCSEvent) error {
	var (
		repoID                uuid.UUID
		isPullRequest         bool
		branch, defaultBranch string
		sha                   string
	)

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

	s.Info("spawning run", "repo_id", repoID)

	workspaces, err := s.ListWorkspacesByRepoID(ctx, repoID)
	if err != nil {
		return err
	}
	if len(workspaces) == 0 {
		s.Info("no connected workspaces found")
		return nil
	}

	// we have 1+ workspaces connected to this repo but we only need to retrieve
	// the repo once, and to do so we'll use the VCS provider associated with
	// the first workspace (any would do).
	if workspaces[0].Connection == nil {
		return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID)
	}
	providerID := workspaces[0].Connection.VCSProviderID

	client, err := s.GetVCSClient(ctx, providerID)
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

	// Determine which workspaces to spawn runs for. If it's a PR then a
	// (speculative) run is spawned for all workspaces. Otherwise each
	// workspace's branch setting is checked and if it set then it must match
	// the event's branch. If it is not set then the event's branch must match
	// the repo's default branch. If neither of these conditions are true then
	// the workspace is skipped.
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
		cv, err := s.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{
			Speculative: internal.Bool(isPullRequest),
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
		if err := s.UploadConfig(ctx, cv.ID, tarball); err != nil {
			return err
		}
		_, err = s.CreateRun(ctx, ws.ID, RunCreateOptions{
			ConfigurationVersionID: internal.String(cv.ID),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
