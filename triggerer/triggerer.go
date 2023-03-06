// Package triggerer handles triggering things in response to incoming VCS
// events.
package triggerer

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/module"
)

// Triggerer triggers jobs in response to incoming VCS events.
type Triggerer struct {
	otf.Application
	*module.Publisher
	logr.Logger

	events <-chan cloud.VCSEvent
}

func NewTriggerer(app otf.Application, logger logr.Logger, events <-chan cloud.VCSEvent) *Triggerer {
	return &Triggerer{
		Application: app,
		Publisher:   module.NewPublisher(app),
		Logger:      logger.WithValues("component", "triggerer"),
		events:      events,
	}
}

// Start handling VCS events and triggering jobs
func (h *Triggerer) Start(ctx context.Context) {
	h.V(2).Info("started")

	for {
		select {
		case event := <-h.events:
			if err := h.handle(ctx, event); err != nil {
				h.Error(err, "handling vcs event")
			}
		case <-ctx.Done():
			return
		}
	}
}

// handle triggers a run upon receiving an event
func (h *Triggerer) handle(ctx context.Context, event cloud.VCSEvent) error {
	if err := h.triggerRun(ctx, event); err != nil {
		return err
	}

	if err := h.PublishFromEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

// triggerRun triggers a run upon receipt of a vcs event
func (h *Triggerer) triggerRun(ctx context.Context, event cloud.VCSEvent) error {
	var webhookID uuid.UUID
	var isPullRequest bool
	var identifier string
	var branch, defaultBranch string
	var sha string

	switch event := event.(type) {
	case cloud.VCSPushEvent:
		webhookID = event.WebhookID
		identifier = event.Identifier
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
		webhookID = event.WebhookID
		identifier = event.Identifier
		sha = event.CommitSHA
		branch = event.Branch
		defaultBranch = event.DefaultBranch
		isPullRequest = true
	}

	h.Info("triggering run", "hook", webhookID)

	workspaces, err := h.ListWorkspacesByWebhookID(ctx, webhookID)
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
	if workspaces[0].Repo() == nil {
		return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID())
	}
	providerID := workspaces[0].Repo().VCSProviderID

	client, err := h.GetVCSClient(ctx, providerID)
	if err != nil {
		return err
	}
	tarball, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Identifier: identifier,
		Ref:        &sha,
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
			if ws.Branch() != "" && ws.Branch() == branch {
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
		if ws.Repo() == nil {
			// Should never happen...
			return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID())
		}
		cv, err := h.CreateConfigurationVersion(ctx, ws.ID(), otf.ConfigurationVersionCreateOptions{
			Speculative: otf.Bool(isPullRequest),
			IngressAttributes: &otf.IngressAttributes{
				// ID     string
				Branch: branch,
				// CloneURL          string
				// CommitMessage     string
				CommitSHA: sha,
				// CommitURL         string
				// CompareURL        string
				Identifier:      identifier,
				IsPullRequest:   isPullRequest,
				OnDefaultBranch: branch == defaultBranch,
			},
		})
		if err != nil {
			return err
		}
		if err := h.UploadConfig(ctx, cv.ID(), tarball); err != nil {
			return err
		}
		_, err = h.CreateRun(ctx, ws.ID(), otf.RunCreateOptions{
			ConfigurationVersionID: otf.String(cv.ID()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
