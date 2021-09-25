package http

// KVStore implementations provide a key-value store.
type KVStore interface {
	Save(key, value string) error
	Load(key string) (value string, err error)
}
