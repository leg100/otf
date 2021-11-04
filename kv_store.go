package otf

var _ KVStore = (*MemStore)(nil)

// KVStore is a key-value store
type KVStore interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte) error
	Delete(key string) error
}

// MemStore is an in-memory key-value store
type MemStore map[string][]byte

func NewMemStore() *MemStore {
	return &MemStore{}
}

func (c MemStore) Get(key string) ([]byte, error) {
	val, ok := map[string][]byte(c)[key]
	if !ok {
		return nil, ErrResourceNotFound
	}

	return val, nil
}

func (c *MemStore) Put(key string, data []byte) error {
	map[string][]byte(*c)[key] = data

	return nil
}

func (c *MemStore) Delete(key string) error {
	delete(map[string][]byte(*c), key)

	return nil
}
