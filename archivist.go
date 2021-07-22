package ots

// Archivist implementations provide a persistent store from and to which binary
// objects can be fetched and uploaded.
type Archivist interface {
	// Get fetches a blob with the given ID
	Get(id string) ([]byte, error)

	// Put uploads a blob and returns an ID uniquely identifying the blob
	Put(blob []byte) (string, error)
}
