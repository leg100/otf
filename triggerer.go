package otf

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

// Triggerer triggers runs in response to incoming VCS events.
type Triggerer struct {
	Application
	logr.Logger

	events <-chan VCSEvent
}

func NewTriggerer(app Application, logger logr.Logger, events <-chan VCSEvent) *Triggerer {
	return &Triggerer{
		Application: app,
		Logger:      logger.WithValues("component", "triggerer"),
		events:      events,
	}
}

// Start handling VCS events and triggering runs
func (h *Triggerer) Start(ctx context.Context) {
	h.V(2).Info("started")

	for {
		select {
		case event := <-h.events:
			h.Info("handling event")
			if err := h.handle(ctx, event); err != nil {
				h.Error(err, "handling event")
			}
		case <-ctx.Done():
			return
		}
	}
}

// handle triggers a run upon receiving an event
func (h *Triggerer) handle(ctx context.Context, event VCSEvent) error {
	// Ignore events for non-default branches that are not a PR.
	if !event.OnDefaultBranch && !event.IsPullRequest {
		h.Info("skipping event on non-default branch")
		return nil
	}

	workspaces, err := h.ListWorkspacesByWebhookID(ctx, event.WebhookID)
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
	providerID := workspaces[0].Repo().ProviderID

	tarball, err := h.GetRepoTarball(ctx, providerID, GetRepoTarballOptions{
		Identifier: event.Identifier,
		Ref:        event.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("retrieving repository tarball: %w", err)
	}

	// create a config version for each workspace and trigger run.
	for _, ws := range workspaces {
		if ws.Repo() == nil {
			return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID())
		}

		cv, err := h.CreateConfigurationVersion(ctx, ws.ID(), ConfigurationVersionCreateOptions{
			Speculative: Bool(event.Branch != ws.Repo().Branch),
			IngressAttributes: &IngressAttributes{
				// ID     string
				Branch: event.Branch,
				// CloneURL          string
				// CommitMessage     string
				CommitSHA: event.CommitSHA,
				// CommitURL         string
				// CompareURL        string
				Identifier:      event.Identifier,
				IsPullRequest:   event.IsPullRequest,
				OnDefaultBranch: event.OnDefaultBranch,
			},
		})
		if err != nil {
			return err
		}
		if err := h.UploadConfig(ctx, cv.ID(), tarball); err != nil {
			return err
		}
		_, err = h.CreateRun(ctx, WorkspaceSpec{ID: String(ws.ID())}, RunCreateOptions{
			ConfigurationVersionID: String(cv.ID()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
