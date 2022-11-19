package otf

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

// VCSEvent is an event received from a VCS provider, e.g. a commit event from
// github
type VCSEvent struct {
	OrganizationName string
	WorkspaceName    string
	// Repo identifier, <owner>/<repo>
	Identifier string
	Branch     string
}

// VCSEventHandler is responsible for handling incoming events from VCS
// providers and performing actions accordingly
type VCSEventHandler struct {
	Application
	logr.Logger

	events <-chan VCSEvent
}

func NewVCSEventHandler(app Application, logger logr.Logger, events <-chan VCSEvent) *VCSEventHandler {
	return &VCSEventHandler{
		Application: app,
		Logger:      logger.WithValues("component", "vcs_event_handler"),
		events:      events,
	}
}

// Start handling VCS events and triggering runs
func (h VCSEventHandler) Start(ctx context.Context) {
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
func (h VCSEventHandler) handle(ctx context.Context, event VCSEvent) error {
	spec := WorkspaceSpec{
		OrganizationName: String(event.OrganizationName),
		Name:             String(event.WorkspaceName),
	}
	ws, err := h.GetWorkspace(ctx, spec)
	if err != nil {
		return err
	}
	if ws.VCSRepo() == nil {
		return fmt.Errorf("workspace is not connected to repo")
	}
	provider, err := h.GetVCSProvider(ctx, ws.VCSRepo().ProviderID, ws.OrganizationName())
	if err != nil {
		return err
	}
	client, err := provider.NewClient(ctx)
	if err != nil {
		return err
	}
	tarball, err := client.GetRepoTarball(ctx, ws.VCSRepo())
	if err != nil {
		return fmt.Errorf("retrieving repository tarball: %w", err)
	}
	cv, err := h.CreateConfigurationVersion(ctx, ws.ID(), ConfigurationVersionCreateOptions{})
	if err != nil {
		return err
	}
	if err := h.UploadConfig(ctx, cv.ID(), tarball); err != nil {
		return err
	}
	_, err = h.CreateRun(ctx, spec, RunCreateOptions{
		ConfigurationVersionID: String(cv.ID()),
	})
	if err != nil {
		return err
	}
	return nil
}
