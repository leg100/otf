package notifications

import (
	"context"
	"errors"
	"sync"
)

type (
	// cache both:
	// (i) efficiently look up of configs
	// (ii) allows re-use of clients whilst ensuring they are closed when no
	// longer in use.
	cache struct {
		mu      sync.Mutex
		clients map[string]*clientEntry // keyed by url
		configs map[string]*Config      // keyed by config ID
	}
	// clientEntry is a record of the number of configs referencing a client.
	clientEntry struct {
		count int
		client
	}
)

// newCache populates a new cache with existing configs
func newCache(ctx context.Context, svc NotificationService) (*cache, error) {
	configs, err := svc.ListNotificationConfigurations(ctx, 
}

// add a config to the cache and either create a client or re-use existing one.
func (c *cache) add(cfg *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.configs[cfg.ID]; ok {
		// this should never happen
		return errors.New("config already added")
	}
	if ent, ok := c.clients[cfg.URL]; ok {
		// re-use existing client
		ent.count++
		c.clients[cfg.URL] = ent
		c.configs[cfg.ID] = cfg
		return nil
	}
	// new cfg; new client
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	c.clients[cfg.URL] = &clientEntry{client: client, count: 1}
	c.configs[cfg.ID] = cfg
	return nil
}

func (c *cache) remove(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, ok := c.configs[id]
	if !ok {
		// this should never happen
		return errors.New("config not found")
	}
	ent, ok := c.clients[cfg.URL]
	if !ok {
		// this should never happen
		return errors.New("client not found")
	}
	ent.count--
	if ent.count == 0 {
		// no more configs reference this client so close and delete
		ent.client.Close()
		delete(c.clients, cfg.URL)
	} else {
		c.clients[cfg.URL] = ent
	}
	delete(c.configs, cfg.ID)
	return nil
}
