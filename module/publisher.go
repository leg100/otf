package module

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/semver"
	"github.com/leg100/otf/vcsprovider"
)

// Publisher publishes new versions of terraform modules from VCS tags
type Publisher struct {
	logr.Logger
	otf.Subscriber
	vcsprovider.VCSProviderService

	db *pgdb

	service
}

// Start handling VCS events and create module versions for new VCS tags
func (h *Publisher) Start(ctx context.Context) error {
	h.V(2).Info("started")

	sub, err := h.Subscribe(ctx, "module-publisher")
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
			if err := h.handleEvent(ctx, event.Payload); err != nil {
				h.Error(err, "handling vcs event")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// PublishFromEvent publishes a module version in response to a vcs event.
func (p *Publisher) handleEvent(ctx context.Context, event cloud.VCSEvent) error {
	// only publish when new tagEvent is created
	tagEvent, ok := event.(cloud.VCSTagEvent)
	if !ok {
		return nil
	}
	if tagEvent.Action != cloud.VCSTagEventCreatedAction {
		return nil
	}
	// only interested in tags that look like semantic versions
	if !semver.IsValid(tagEvent.Tag) {
		return nil
	}
	module, err := p.GetModuleByRepoID(ctx, tagEvent.RepoID)
	if err != nil {
		return err
	}
	if module.Connection == nil {
		return fmt.Errorf("module is not connected to a repo: %s", module.ID)
	}
	client, err := p.GetVCSClient(ctx, module.Connection.VCSProviderID)
	if err != nil {
		return err
	}
	return p.PublishVersion(ctx, PublishModuleVersionOptions{
		ModuleID: module.ID,
		// strip off v prefix if it has one
		Version: strings.TrimPrefix(tagEvent.Tag, "v"),
		Ref:     tagEvent.CommitSHA,
		Repo:    repo(module.Connection.Repo),
		Client:  client,
	})
}
