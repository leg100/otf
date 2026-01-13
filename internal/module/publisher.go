package module

import (
	"context"
	"fmt"
	"strings"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/vcs"
)

type (
	// publisher publishes new versions of terraform modules from VCS tags
	publisher struct {
		logr.Logger

		modules      *Service
		vcsproviders *vcs.Service
	}
)

func (p *publisher) handle(event vcs.Event) {
	logger := p.Logger.WithValues(
		"sha", event.CommitSHA,
		"type", event.Type,
		"action", event.Action,
		"branch", event.Branch,
		"tag", event.Tag,
		"repo", event.Repo,
	)

	if err := p.handleWithError(event); err != nil {
		logger.Error(err, "handling event")
	}
}

// handlerWithError publishes a module version in response to a vcs event.
func (p *publisher) handleWithError(event vcs.Event) error {
	// no parent context; handler is called asynchronously
	ctx := context.Background()
	// give spawner unlimited powers
	ctx = authz.AddSubjectToContext(ctx, &authz.Superuser{Username: "run-spawner"})

	// only create-tag events trigger the publishing of new module version
	if event.Type != vcs.EventTypeTag {
		return nil
	}
	if event.Action != vcs.ActionCreated {
		return nil
	}
	// only interested in tags that look like semantic versions
	if !semver.IsValid(event.Tag) {
		return nil
	}
	// TODO: we're only retrieving *one* module, but can not *multiple* modules
	// be connected to a repo?
	module, err := p.modules.GetModuleByConnection(ctx, event.VCSProviderID, event.Repo)
	if err != nil {
		return fmt.Errorf("retrieving module: %w", err)
	}
	if module.Connection == nil {
		return fmt.Errorf("module is not connected to a repo: %s", module.ID)
	}
	client, err := p.vcsproviders.Get(ctx, module.Connection.VCSProviderID)
	if err != nil {
		return fmt.Errorf("retrieving vcs provider: %w", err)
	}
	return p.modules.PublishVersion(ctx, PublishVersionOptions{
		ModuleID: module.ID,
		// strip off v prefix if it has one
		Version: strings.TrimPrefix(event.Tag, "v"),
		Ref:     event.CommitSHA,
		Repo:    Repo(module.Connection.Repo),
		Client:  client,
	})
}
