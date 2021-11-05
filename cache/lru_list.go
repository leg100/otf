package cache

import (
	"container/list"
	"fmt"
)

func NewLRUList() *LRUList {
	l := LRUList{
		List:   list.New(),
		lookup: make(map[string]*list.Element),
	}

	return &l
}

// LRUList maintains a list of least recently used keys (strings).
type LRUList struct {
	// doubly-linked list for evicting keys
	*list.List

	// ... and maintain another data structure for looking up keys.
	lookup map[string]*list.Element
}

func (l *LRUList) Add(key string) {
	val := l.PushFront(key)

	l.lookup[key] = val
}

// Evict removes the key from the list.
func (l *LRUList) Evict(key string) error {
	val, ok := l.lookup[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	l.Remove(val)

	return nil
}

// Oldest returns the least recently used key. Nil means the key could not be
// found.
func (l *LRUList) Oldest() *string {
	val := l.Back().Value
	if val == nil {
		return nil
	}

	s := val.(string)
	return &s
}

// Freshen marks the key as recently used.
func (l *LRUList) Freshen(key string) error {
	val, ok := l.lookup[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	l.MoveToFront(val)

	return nil
}
