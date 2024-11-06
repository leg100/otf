package notifications

import (
	"context"
	"errors"
	"sync"

	"github.com/leg100/otf/internal/resource"
)

type (
	// cache both:
	// (i) efficiently look up of configs
	// (ii) allows re-use of clients whilst ensuring they are closed when no
	// longer in use.
	//
	// A client is maintained per unique url.
	cache struct {
		mu      sync.Mutex
		clients map[string]*clientEntry // keyed by url
		configs map[resource.ID]*Config // keyed by config ID

		clientFactory // constructs new clients
	}
	// clientEntry is a record of the number of configs referencing a client.
	clientEntry struct {
		count int
		client
	}
	// a db from which to populate the cache
	cacheDB interface {
		listAll(context.Context) ([]*Config, error)
	}
)

// newCache populates a new cache with existing configs
func newCache(ctx context.Context, db cacheDB, f clientFactory) (*cache, error) {
	configs, err := db.listAll(ctx)
	if err != nil {
		return nil, err
	}
	cache := &cache{
		configs:       make(map[resource.ID]*Config, len(configs)),
		clients:       make(map[string]*clientEntry),
		clientFactory: f,
	}

	for _, cfg := range configs {
		if err := cache.add(cfg); err != nil {
			return nil, err
		}
	}
	return cache, nil
}

// add a config to the cache and either create a client or re-use existing one.
func (c *cache) add(cfg *Config) error {
	if cfg.DestinationType == DestinationEmail {
		// email type is unimplemented
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.configs[cfg.ID]; ok {
		// this should never happen
		return errors.New("config already added")
	}
	if ent, ok := c.clients[*cfg.URL]; ok {
		// re-use existing client
		ent.count++
		c.clients[*cfg.URL] = ent
		c.configs[cfg.ID] = cfg
		configsMetric.Inc()
		return nil
	}
	// new cfg; new client
	client, err := c.newClient(cfg)
	if err != nil {
		return err
	}
	c.clients[*cfg.URL] = &clientEntry{client: client, count: 1}
	clientsMetric.Inc()
	c.configs[cfg.ID] = cfg
	configsMetric.Inc()
	return nil
}

func (c *cache) remove(id resource.ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, ok := c.configs[id]
	if !ok {
		// this should never happen
		return errors.New("config not found")
	}
	ent, ok := c.clients[*cfg.URL]
	if !ok {
		// this should never happen
		return errors.New("client not found")
	}
	ent.count--
	if ent.count == 0 {
		// no more configs reference this client so close and delete
		ent.Close()
		delete(c.clients, *cfg.URL)
		clientsMetric.Dec()
	} else {
		c.clients[*cfg.URL] = ent
	}
	delete(c.configs, cfg.ID)
	configsMetric.Dec()
	return nil
}
