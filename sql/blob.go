package sql

import (
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.BlobStore = (*BlobDB)(nil)
)

type BlobDB struct {
	*sqlx.DB
}

func NewBlobDB(db *sqlx.DB) *BlobDB {
	return &BlobDB{
		DB: db,
	}
}

// Put persists a Blob to the DB.
func (db BlobDB) Put(id string, blob []byte) error {
	sql := "INSERT INTO blobs (blob_id, blob) VALUES ($1, $2)"

	_, err := db.Exec(sql, id, blob)
	if err != nil {
		return err
	}

	return nil
}

func (db BlobDB) Get(id string) ([]byte, error) {
	var sql = "SELECT blob FROM blobs WHERE blob_id = $1"
	var blob []byte

	if err := db.DB.Get(&blob, sql, id); err != nil {
		return nil, databaseError(err)
	}

	return blob, nil
}
