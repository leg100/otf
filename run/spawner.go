package run

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

// spawner spawns new runs in response to vcs events
type spawner struct {
	logr.Logger
	otf.Subscriber
	otf.ConfigurationVersionService
	otf.RepoService
	otf.WorkspaceService
	otf.VCSProviderService

	service
}

// Start handling VCS events and triggering jobs
func (h *spawner) Start(ctx context.Context) error {
	sub, err := h.Subscribe(ctx, "run-spawner")
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-sub:
			// skip non-vcs events
			if event.Type != otf.EventVCS {
				continue
			}

			if err := h.handle(ctx, event.Payload); err != nil {
				h.Error(err, "handling vcs event")
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
		// only trigger runs when opening a PR and pushing to a PR
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

	h.Info("triggering run", "repo_id", repoID)

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

	// Determine which workspaces to trigger runs for. If its a PR then a
	// (speculative) run is
	// triggered for all workspaces. Otherwise each workspace's branch setting
	// is checked and if it set then it must match the event's branch. If it is
	// not set then the event's branch must match the repo's default branch. If
	// neither of these conditions are true then the workspace is skipped.
	filterFunc := func(unfiltered []*otf.Workspace) (filtered []*otf.Workspace) {
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

	// create a config version for each workspace and trigger run.
	for _, ws := range workspaces {
		if ws.Connection == nil {
			// Should never happen...
			return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID)
		}
		cv, err := h.CreateConfigurationVersion(ctx, ws.ID, otf.ConfigurationVersionCreateOptions{
			Speculative: otf.Bool(isPullRequest),
			IngressAttributes: &otf.IngressAttributes{
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
		_, err = h.create(ctx, ws.ID, RunCreateOptions{
			ConfigurationVersionID: otf.String(cv.ID),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
