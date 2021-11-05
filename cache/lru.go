package cache

import (
	"fmt"
	"sync"
)

// LRU provides a least recently used cache for binary objects, bounded by a
// maximum size.
type LRU struct {
	// KVStore is the underlying key-value store for the binary objects
	KVStore

	// LRUList lists least recently used entries for eviction
	*LRUList

	// capacity is max size of cache in bytes
	capacity int

	mu sync.Mutex

	// usage is number of bytes the cache is currently using
	usage int
}

// NewLRUMemstore constructs a LRU cache backed by memory.
func NewLRUMemstore(capacity int) *LRU {
	c := LRU{
		capacity: capacity,
		LRUList:  NewLRUList(),
		KVStore:  NewMemstore(),
	}

	return &c
}

func (c *LRU) Get(key string) ([]byte, error) {
	val, err := c.KVStore.Get(key)
	if err != nil {
		return nil, err
	}

	if err := c.Freshen(key); err != nil {
		return nil, err
	}

	return val, nil
}

func (c *LRU) Put(key string, data []byte) error {
	if len(data) > c.capacity {
		return fmt.Errorf("cannot cache item bigger than cache capacity: %d > %d", len(data), c.capacity)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		if c.usage+len(data) > c.capacity {
			evicted, err := c.evict()
			if err != nil {
				return err
			}
			c.usage -= len(evicted)
		} else {
			break
		}
	}

	if err := c.KVStore.Put(key, data); err != nil {
		return err
	}

	c.LRUList.Add(key)

	c.usage += len(data)

	return nil
}

func (c *LRU) evict() ([]byte, error) {
	key := c.LRUList.Oldest()
	if key == nil {
		return nil, fmt.Errorf("cannot evict; LRU evictor is empty")
	}

	val, err := c.KVStore.Get(*key)
	if err != nil {
		return nil, err
	}

	if err := c.KVStore.Delete(*key); err != nil {
		return nil, err
	}

	if err := c.LRUList.Evict(*key); err != nil {
		return nil, err
	}

	return val, nil
}
