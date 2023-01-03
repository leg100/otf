package otf

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

// Triggerer triggers jobs in response to incoming VCS events.
type Triggerer struct {
	Application
	*Publisher
	logr.Logger

	events <-chan VCSEvent
}

func NewTriggerer(app Application, logger logr.Logger, events <-chan VCSEvent) *Triggerer {
	return &Triggerer{
		Application: app,
		Publisher:   NewPublisher(app),
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
				h.Error(err, "handling event")
			}
		case <-ctx.Done():
			return
		}
	}
}

// handle triggers a run upon receiving an event
func (h *Triggerer) handle(ctx context.Context, event VCSEvent) error {
	if err := h.triggerRun(ctx, event); err != nil {
		return err
	}

	if err := h.PublishFromEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

// triggerRun triggers a run upon receipt of a vcs vevent
func (h *Triggerer) triggerRun(ctx context.Context, event VCSEvent) error {
	var webhookID uuid.UUID
	var isPullRequest bool
	var identifier string
	var branch string
	var sha string

	switch event := event.(type) {
	case *VCSPushEvent:
		webhookID = event.WebhookID
		identifier = event.Identifier
		sha = event.CommitSHA
		branch = event.Branch
	case *VCSPullEvent:
		if event.Action != VCSPullEventUpdated {
			// ignore all other pull events
			return nil
		}
		webhookID = event.WebhookID
		identifier = event.Identifier
		sha = event.CommitSHA
		branch = event.Branch
		isPullRequest = true
	}

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
	providerID := workspaces[0].Repo().ProviderID

	tarball, err := h.GetRepoTarball(ctx, providerID, GetRepoTarballOptions{
		Identifier: identifier,
		Ref:        sha,
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
			Speculative: Bool(isPullRequest),
			IngressAttributes: &IngressAttributes{
				// ID     string
				Branch: branch,
				// CloneURL          string
				// CommitMessage     string
				CommitSHA: sha,
				// CommitURL         string
				// CompareURL        string
				Identifier:      identifier,
				IsPullRequest:   isPullRequest,
				OnDefaultBranch: (ws.Repo().Branch == branch),
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
