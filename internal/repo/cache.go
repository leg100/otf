package repo

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	cache struct {
		// keyed by vcs provider ID
		providers map[string]*vcsprovider.VCSProvider
		// keyed by hook ID
		hooks map[uuid.UUID]*hook

		mu sync.Mutex
	}

	// cache constructor options
	cacheOptions struct {
		vcsprovider.VCSProviderService
		pubsub.Subscriber
		hookdb *db
	}
)

func newCache(ctx context.Context, opts cacheOptions) (*cache, error) {
	providers, err := opts.VCSProviderService.ListAllVCSProviders(ctx)
	if err != nil {
		return nil, err
	}
	hooks, err := opts.hookdb.listHooks(ctx)
	if err != nil {
		return nil, err
	}
	cache := &cache{
		providers: make(map[string]*vcsprovider.VCSProvider, len(providers)),
		hooks:     make(map[uuid.UUID]*hook, len(hooks)),
	}
	for _, h := range hooks {
		cache.hooks[h.id] = h
	}
	for _, p := range providers {
		cache.providers[p.ID] = p
	}
	return cache, nil
}

func (c *cache) getHook(hookID uuid.UUID) *hook {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.hooks[hookID]
}

func (c *cache) getProvider(providerID string) *vcsprovider.VCSProvider {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.providers[providerID]
}

func (c *cache) setHook(hook *hook) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.hooks[hook.id] = hook
}

func (c *cache) setProvider(provider *vcsprovider.VCSProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.providers[provider.ID] = provider
}

func (c *cache) delete(hookID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hook, ok := c.hooks[hookID]
	if !ok {
		// hook not found in cache, return without error
		return
	}
	delete(c.hooks, hookID)

	// delete webhook's vcs provider if no longer referenced by any other
	// webhook
	for _, h := range c.hooks {
		if h.vcsProviderID == hook.vcsProviderID {
			// another webhook is referencing the vcs provider
			return
		}
	}
	// deleted unreferenced provider
	delete(c.providers, hook.vcsProviderID)
}
