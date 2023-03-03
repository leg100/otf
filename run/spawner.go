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
	var webhookID uuid.UUID
	var isPullRequest bool
	var identifier string
	var branch string
	var sha string

	switch event := event.(type) {
	case cloud.VCSPushEvent:
		webhookID = event.WebhookID
		identifier = event.Identifier
		sha = event.CommitSHA
		branch = event.Branch
	case cloud.VCSPullEvent:
		if event.Action != cloud.VCSPullEventUpdated {
			// ignore all other pull events
			return nil
		}
		webhookID = event.WebhookID
		identifier = event.Identifier
		sha = event.CommitSHA
		branch = event.Branch
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
	if workspaces[0].Repo == nil {
		return fmt.Errorf("workspace is not connected to a repo: %s", workspaces[0].ID)
	}
	providerID := workspaces[0].Repo.ProviderID

	client, err := h.GetVCSClient(ctx, providerID)
	if err != nil {
		return err
	}
	tarball, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Identifier: identifier,
		Ref:        sha,
	})
	if err != nil {
		return fmt.Errorf("retrieving repository tarball: %w", err)
	}

	// create a config version for each workspace and trigger run.
	for _, ws := range workspaces {
		if ws.Repo == nil {
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
				Identifier:      identifier,
				IsPullRequest:   isPullRequest,
				OnDefaultBranch: (ws.Repo.Branch == branch),
			},
		})
		if err != nil {
			return err
		}
		if err := h.UploadConfig(ctx, cv.ID, tarball); err != nil {
			return err
		}
		_, err = h.create(ctx, ws.ID, otf.RunCreateOptions{
			ConfigurationVersionID: otf.String(cv.ID),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
