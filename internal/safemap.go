package internal

import (
	"sync"
)

type SafeMap[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m: make(map[K]V),
	}
}

func (r *SafeMap[K, V]) Set(key K, value V) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.m[key] = value
}

func (r *SafeMap[K, V]) Get(key K) (V, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.m[key]
	return value, ok
}
