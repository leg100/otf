package main

import "github.com/leg100/otf/http"

var _ http.KVStore = (KVMap)(nil)

// KVMap is a basic implementation of http.KVStore for testing purposes.
type KVMap map[string]string

func (m KVMap) Save(key, value string) error {
	m[key] = value
	return nil
}

func (m KVMap) Load(key string) (string, error) {
	return m[key], nil
}
