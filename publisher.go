package otf

import (
	"context"
	"fmt"
	"strings"

	"github.com/leg100/otf/semver"
	"github.com/pkg/errors"
)

// Publisher publishes terraform modules.
type Publisher struct {
	Application
	*WebhookCreator
	*WebhookUpdater
}

// PublishModule publishes a new module from a VCS repository, enumerating through
// its git tags and releasing a module version for each tag.
func (p *Publisher) PublishModule(ctx context.Context, opts PublishModuleOptions) (*Module, error) {
	repo, err := p.GetRepository(ctx, opts.ProviderID, opts.Identifier)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving repository info")
	}

	vcsProvider, err := p.GetVCSProvider(ctx, opts.ProviderID)
	if err != nil {
		return nil, err
	}

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

	var mod *Module
	err = p.Tx(ctx, func(app Application) (err error) {
		webhook, err := app.DB().SyncWebhook(ctx, SyncWebhookOptions{
			Identifier:        opts.Identifier,
			HTTPURL:           repo.HTTPURL,
			ProviderID:        opts.ProviderID,
			OTFHost:           opts.OTFHost,
			Cloud:             vcsProvider.cloudConfig.Name,
			CreateWebhookFunc: p.Create,
			UpdateWebhookFunc: p.Update,
		})
		if err != nil {
			return errors.Wrap(err, "syncing webhook")
		}

		mod = NewModule(CreateModuleOptions{
			Name:         name,
			Provider:     provider,
			Organization: opts.Organization,
			Repo: &ModuleRepo{
				WebhookID:  webhook.WebhookID,
				ProviderID: opts.ProviderID,
				HTTPURL:    repo.HTTPURL,
				Identifier: repo.Identifier,
			},
		})

		err = app.DB().CreateModule(ctx, mod)
		if err != nil {
			return errors.Wrap(err, "persisting module")
		}

		// Make new version for each tag that looks like a semantic version.
		tags, err := p.ListTags(ctx, opts.ProviderID, ListTagsOptions{
			Identifier: opts.Identifier,
		})
		if err != nil {
			return err
		}
		for _, tag := range tags {
			_, version, found := strings.Cut(string(tag), "/")
			if !found {
				return fmt.Errorf("malformed git ref: %s", tag)
			}

			// skip tags that are not semantic versions
			if !semver.IsValid(version) {
				continue
			}

			modVersion, err := p.PublishVersion(ctx, PublishModuleVersionOptions{
				ModuleID: mod.ID(),
				// strip off v prefix if it has one
				Version:    strings.TrimPrefix(version, "v"),
				Ref:        string(tag),
				Identifier: opts.Identifier,
				ProviderID: mod.Repo().ProviderID,
			})
			if err != nil {
				return err
			}
			mod.versions = append(mod.versions, modVersion)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return mod, nil
}

// Publish a module version in response to a vcs event.
func (p *Publisher) PublishFromEvent(ctx context.Context, event VCSEvent) error {
	// only publish when new tag is created
	tag, ok := event.(*VCSTagEvent)
	if !ok {
		return nil
	}
	if tag.Action != VCSTagEventCreatedAction {
		return nil
	}
	// only interested in tags that look like semantic versions
	if !semver.IsValid(tag.Tag) {
		return nil
	}

	module, err := p.GetModuleByWebhookID(ctx, tag.WebhookID)
	if err != nil {
		return err
	}
	if module.Repo() == nil {
		return fmt.Errorf("module is not connected to a repo: %s", module.ID())
	}

	// skip older or equal versions
	currentVersion := module.LatestVersion().Version()
	if n := semver.Compare(tag.Tag, currentVersion); n <= 0 {
		return nil
	}

	_, err = p.PublishVersion(ctx, PublishModuleVersionOptions{
		ModuleID: module.ID(),
		// strip off v prefix if it has one
		Version:    strings.TrimPrefix(tag.Tag, "v"),
		Ref:        tag.CommitSHA,
		Identifier: tag.Identifier,
		ProviderID: module.Repo().ProviderID,
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

// Publish a module version, retrieving its contents from a repository and
// uploading it to the module store.
func (p *Publisher) PublishVersion(ctx context.Context, opts PublishModuleVersionOptions) (*ModuleVersion, error) {
	version, err := p.CreateModuleVersion(ctx, CreateModuleVersionOptions{
		ModuleID: opts.ModuleID,
		Version:  opts.Version,
	})
	if err != nil {
		return nil, err
	}

	tarball, err := p.GetRepoTarball(ctx, opts.ProviderID, GetRepoTarballOptions{
		Identifier: opts.Identifier,
		Ref:        opts.Ref,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving repository tarball: %w", err)
	}

	err = p.UploadModuleVersion(ctx, UploadModuleVersionOptions{
		ModuleVersionID: version.ID(),
		Tarball:         tarball,
	})
	if err != nil {
		return nil, err
	}
	return version, nil
}
