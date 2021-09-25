package main

var _ KVStore = (KVMap)(nil)

// KVMap is a basic implementation of KVStore for testing purposes.
type KVMap map[string]string

func (m KVMap) Save(key, value string) error {
	m[key] = value
	return nil
}

func (m KVMap) Load(key string) (string, error) {
	return m[key], nil
}
