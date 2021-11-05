package cache

import "github.com/leg100/otf"

var _ KVStore = (*Memstore)(nil)

// Memstore is an in-memory key-value store.
type Memstore map[string][]byte

func NewMemstore() *Memstore {
	return &Memstore{}
}

func (c Memstore) Get(key string) ([]byte, error) {
	val, ok := map[string][]byte(c)[key]
	if !ok {
		return nil, otf.ErrResourceNotFound
	}

	return val, nil
}

func (c *Memstore) Put(key string, data []byte) error {
	map[string][]byte(*c)[key] = data

	return nil
}

func (c *Memstore) Delete(key string) error {
	delete(map[string][]byte(*c), key)

	return nil
}
