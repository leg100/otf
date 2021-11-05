package cache

// KVStore is a key-value store
type KVStore interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte) error
	Delete(key string) error
}
