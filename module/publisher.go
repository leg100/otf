package module

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/semver"
)

// Publisher publishes terraform modules.
type Publisher struct {
	otf.RepoService // for registering and unregistering connections to webhooks
	otf.Subscriber
	otf.VCSProviderService
	logr.Logger

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

			if err := h.PublishFromEvent(ctx, event.Payload); err != nil {
				h.Error(err, "handling vcs event")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// PublishModule publishes a new module from a VCS repository, enumerating through
// its git tags and releasing a module version for each tag.
func (p *Publisher) PublishModule(ctx context.Context, opts PublishModuleOptions) (*Module, error) {
	_, repoName, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}
	parts := strings.SplitN(repoName, "-", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid repository name: %s", repoName)
	}
	provider := parts[1]
	name := parts[2]

	mod := newModule(CreateModuleOptions{
		Name:         name,
		Provider:     provider,
		Organization: opts.Organization,
	})

	// persist module to db and connect mod to repo
	err := p.db.tx(ctx, func(tx *pgdb) error {
		if err := tx.CreateModule(ctx, mod); err != nil {
			return err
		}
		connection, err := p.RepoService.Connect(ctx, otf.ConnectOptions{
			ConnectionType: otf.ModuleConnection,
			ResourceID:     mod.id,
			VCSProviderID:  opts.ProviderID,
			Identifier:     opts.Identifier,
			Tx:             tx,
		})
		if err != nil {
			return err
		}
		mod.repo = connection
		return nil
	})
	if err != nil {
		return nil, err
	}

	client, err := p.GetVCSClient(ctx, opts.ProviderID)
	if err != nil {
		return nil, err
	}

	// Make new version for each tag that looks like a semantic version.
	tags, err := client.ListTags(ctx, cloud.ListTagsOptions{
		Identifier: opts.Identifier,
	})
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return p.UpdateModuleStatus(ctx, UpdateModuleStatusOptions{
			ID:     mod.id,
			Status: ModuleStatusNoVersionTags,
		})
	}

	for _, tag := range tags {
		// tags/<version> -> <version>
		_, version, found := strings.Cut(tag, "/")
		if !found {
			return nil, fmt.Errorf("malformed git ref: %s", tag)
		}

		// skip tags that are not semantic versions
		if !semver.IsValid(version) {
			continue
		}

		mod, _, err = p.PublishVersion(ctx, PublishModuleVersionOptions{
			ModuleID: mod.id,
			// strip off v prefix if it has one
			Version:    strings.TrimPrefix(version, "v"),
			Ref:        tag,
			Identifier: opts.Identifier,
			ProviderID: opts.ProviderID,
		})
		if err != nil {
			return nil, err
		}
	}
	return mod, nil
}

// PublishFromEvent publishes a module version in response to a vcs event.
func (p *Publisher) PublishFromEvent(ctx context.Context, event cloud.VCSEvent) error {
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

	module, err := p.GetModuleByWebhookID(ctx, tagEvent.WebhookID)
	if err != nil {
		return err
	}
	if module.repo == nil {
		return fmt.Errorf("module is not connected to a repo: %s", module.id)
	}

	// skip older or equal versions
	latestVersion := module.latest.version
	if n := semver.Compare(tagEvent.Tag, latestVersion); n <= 0 {
		return nil
	}

	_, _, err = p.PublishVersion(ctx, PublishModuleVersionOptions{
		ModuleID: module.id,
		// strip off v prefix if it has one
		Version:    strings.TrimPrefix(tagEvent.Tag, "v"),
		Ref:        tagEvent.CommitSHA,
		Identifier: tagEvent.Identifier,
		ProviderID: module.repo.VCSProviderID,
	})
	if err != nil {
		return err
	}

	return nil
}

type PublishModuleVersionOptions struct {
	ModuleID   string
	Version    string
	Ref        string
	Identifier string
	ProviderID string
}

// PublishVersion publishes a module version, retrieving its contents from a repository and
// uploading it to the module store.
func (p *Publisher) PublishVersion(ctx context.Context, opts PublishModuleVersionOptions) (*Module, *ModuleVersion, error) {
	modver, err := p.CreateModuleVersion(ctx, CreateModuleVersionOptions{
		ModuleID: opts.ModuleID,
		Version:  opts.Version,
	})
	if err != nil {
		return nil, nil, err
	}

	client, err := p.GetVCSClient(ctx, opts.ProviderID)
	if err != nil {
		return nil, nil, err
	}

	tarball, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Identifier: opts.Identifier,
		Ref:        &opts.Ref,
	})
	if err != nil {
		return UpdateModuleVersionStatus(ctx, p, UpdateModuleVersionStatusOptions{
			ID:     modver.id,
			Status: ModuleVersionStatusCloneFailed,
			Error:  err.Error(),
		})
	}

	return p.Upload(ctx, UploadModuleVersionOptions{
		VersionID: modver.id,
		Tarball:   tarball,
	})
}
