package module

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	// Publisher publishes new versions of terraform modules from VCS tags
	Publisher struct {
		logr.Logger
		pubsub.Subscriber
		vcsprovider.VCSProviderService
		ModuleService
	}
)

// Start starts handling VCS events and publishing modules accordingly
func (p *Publisher) Start(ctx context.Context) error {
	sub, err := p.Subscribe(ctx, "module-publisher-")
	if err != nil {
		return err
	}

	for event := range sub {
		// skip non-vcs events
		if event.Type != pubsub.EventVCS {
			continue
		}
		if err := p.handleEvent(ctx, event.Payload); err != nil {
			p.Error(err, "handling vcs event")
		}
	}
	return nil
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
	return p.PublishVersion(ctx, PublishVersionOptions{
		ModuleID: module.ID,
		// strip off v prefix if it has one
		Version: strings.TrimPrefix(tagEvent.Tag, "v"),
		Ref:     tagEvent.CommitSHA,
		Repo:    Repo(module.Connection.Repo),
		Client:  client,
	})
}
