package otf

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/mod/semver"
)

type Module struct {
	id           string
	createdAt    time.Time
	updatedAt    time.Time
	name         string
	provider     string
	organization *Organization // Module belongs to an organization
	repo         *ModuleRepo
	versions     []*ModuleVersion
}

type ModuleVersion struct {
	version   string
	createdAt time.Time
	updatedAt time.Time
	content   []byte  // tar.gz
	module    *Module // ModuleVersion belongs to a module
}

type ModuleRepo struct {
	ProviderID string
	WebhookID  uuid.UUID
	Identifier string // identifier is <repo_owner>/<repo_name>
	HTTPURL    string // HTTPURL is the web url for the repo
}

type ModuleService interface {
	CreateModule(context.Context, ModuleCreateOptions) (*Module, error)
	CreateModuleVersion(context.Context, ModuleCreateVersionOptions) (*ModuleVersion, error)
	ListModules(context.Context, ModuleListOptions) (*ModuleList, error)
	GetModule(context.Context, ModuleListOptions) (*Module, error)
}

type (
	ModuleCreateOptions struct {
		Name         string
		Provider     string
		Repo         *ModuleRepo
		Organization *Organization
	}
	ModuleCreateVersionOptions struct {
		ModuleID string
		Tarball  []byte
		Version  string
	}
	ModuleListOptions struct {
		ListOptions
	}
	ModuleList struct {
		*Pagination
		Items []*Module
	}
)

type ModuleMaker struct {
	ModuleService
	VCSProviderService
}

func (mm *ModuleMaker) NewModule(ctx context.Context, opts ModuleCreateOptions) (*Module, error) {
	mod := NewModule(opts)

	if opts.Repo != nil {
		// list all tags starting with 'v' in the module's repo
		tags, err := mm.ListTags(ctx, opts.Repo.ProviderID, ListTagsOptions{
			Identifier: opts.Repo.Identifier,
		})
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			_, version, found := strings.Cut(tag.Ref, "/")
			if !found {
				return nil, fmt.Errorf("malformed git tag ref: %s", tag.Ref)
			}

			// strip off 'v' prefix if it has one
			version = strings.TrimPrefix(version, "v")

			// skip tags that are not semantic versions
			if !semver.IsValid(version) {
				continue
			}
			tarball, err := mm.GetRepoTarball(ctx, opts.Repo.ProviderID, GetRepoTarballOptions{
				Ref: tag.Ref,
			})
			if err != nil {
				return nil, err
			}
			_, err = mm.CreateModuleVersion(ctx, ModuleCreateVersionOptions{
				ModuleID: mod.id,
				Version:  version,
				Tarball:  tarball,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return mod, nil
}

func NewModule(opts ModuleCreateOptions) *Module {
	m := Module{
		id:           NewID("mod"),
		createdAt:    CurrentTimestamp(),
		updatedAt:    CurrentTimestamp(),
		name:         opts.Name,
		provider:     opts.Provider,
		organization: opts.Organization,
	}
	return &m
}
