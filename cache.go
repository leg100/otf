package otf

import (
	"container/list"
	"fmt"
	"sync"
)

// LRUCache provides a least-recently-used cache
type LRUCache struct {
	// KVStore is the underlying key-value store implementation
	KVStore

	// EvictList lists least recently used entries for eviction
	*EvictList

	// capacity is max size of cache in bytes
	capacity int

	mu sync.Mutex

	// usage is number of bytes the cache is currently using
	usage int
}

func NewLRUMemoryCache(capacity int) *LRUCache {
	c := LRUCache{
		capacity:  capacity,
		EvictList: NewEvictList(),
		KVStore:   NewMemStore(),
	}

	return &c
}

func (c *LRUCache) Get(key string) ([]byte, error) {
	val, err := c.KVStore.Get(key)
	if err != nil {
		return nil, err
	}

	if err := c.Freshen(key); err != nil {
		return nil, err
	}

	return val, nil
}

func (c *LRUCache) Put(key string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		if c.usage+len(data) > c.capacity {
			if err := c.evict(); err != nil {
				return err
			}
		} else {
			break
		}
	}

	if err := c.KVStore.Put(key, data); err != nil {
		return err
	}

	c.EvictList.Add(key)

	return nil
}

func (c *LRUCache) evict() error {
	key := c.EvictList.Oldest()
	if key == nil {
		return fmt.Errorf("LRU evictor is empty")
	}

	if err := c.KVStore.Delete(*key); err != nil {
		return err
	}

	if err := c.EvictList.Evict(*key); err != nil {
		return err
	}

	return nil
}

func NewEvictList() *EvictList {
	l := EvictList{
		List:   list.New(),
		lookup: make(map[string]*list.Element),
	}

	return &l
}

type EvictList struct {
	// doubly-linked list for evicting entries
	*list.List
	lookup map[string]*list.Element
}

func (l *EvictList) Add(key string) {
	val := l.PushFront(key)

	l.lookup[key] = val
}

func (l *EvictList) Evict(key string) error {
	val, ok := l.lookup[key]
	if !ok {
		return fmt.Errorf("key not present in evict list: %s", key)
	}

	l.Remove(val)

	return nil
}

func (l *EvictList) Oldest() *string {
	val := l.Back().Value
	if val == nil {
		return nil
	}

	return val.(*string)
}

func (l *EvictList) Freshen(key string) error {
	val, ok := l.lookup[key]
	if !ok {
		return fmt.Errorf("key not present in evict list: %s", key)
	}

	l.MoveToFront(val)

	return nil
}
