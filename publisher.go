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
	*ModuleVersionUploader
	*WebhookCreator
	*WebhookUpdater
}

func NewPublisher(app Application) *Publisher {
	return &Publisher{
		Application: app,
		ModuleVersionUploader: &ModuleVersionUploader{
			Application: app,
		},
		WebhookCreator: &WebhookCreator{
			VCSProviderService: app,
			CloudService:       app,
			HostnameService:    app,
		},
		WebhookUpdater: &WebhookUpdater{
			VCSProviderService: app,
			HostnameService:    app,
		},
	}
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
			ProviderID:        opts.ProviderID,
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
				Identifier: repo.Identifier,
			},
		})

		err = app.DB().CreateModule(ctx, mod)
		if err != nil {
			return errors.Wrap(err, "persisting module")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Make new version for each tag that looks like a semantic version.
	tags, err := p.ListTags(ctx, opts.ProviderID, ListTagsOptions{
		Identifier: opts.Identifier,
	})
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return p.UpdateModuleStatus(ctx, UpdateModuleStatusOptions{
			ID:     mod.ID(),
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
			ModuleID: mod.ID(),
			// strip off v prefix if it has one
			Version:    strings.TrimPrefix(version, "v"),
			Ref:        tag,
			Identifier: opts.Identifier,
			ProviderID: mod.Repo().ProviderID,
		})
		if err != nil {
			return nil, err
		}
	}
	return mod, nil
}

// PublishFromEvent publishes a module version in response to a vcs event.
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
	latestVersion := module.Latest().Version()
	if n := semver.Compare(tag.Tag, latestVersion); n <= 0 {
		return nil
	}

	_, _, err = p.PublishVersion(ctx, PublishModuleVersionOptions{
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

	tarball, err := p.GetRepoTarball(ctx, opts.ProviderID, GetRepoTarballOptions{
		Identifier: opts.Identifier,
		Ref:        opts.Ref,
	})
	if err != nil {
		return UpdateModuleVersionStatus(ctx, p, UpdateModuleVersionStatusOptions{
			ID:     modver.ID(),
			Status: ModuleVersionStatusCloneFailed,
			Error:  err.Error(),
		})
	}

	return p.Upload(ctx, UploadModuleVersionOptions{
		ModuleVersionID: modver.ID(),
		Tarball:         tarball,
	})
}
