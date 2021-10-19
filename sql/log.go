package sql

import (
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.LogStore = (*LogDB)(nil)
)

type LogDB struct {
	*sqlx.DB
}

func NewLogDB(db *sqlx.DB) *LogDB {
	return &LogDB{
		DB: db,
	}
}

// Put persists a Log to the DB.
func (db LogDB) PutChunk(id string, blob []byte, opts otf.PutChunkOptions) error {
	sql := "INSERT INTO logs (external_id, blob, start, _end) VALUES (?, ?, ?, ?)"

	_, err := db.Exec(sql, id, blob, opts.Start, opts.End)
	if err != nil {
		return err
	}

	return nil
}

func (db LogDB) GetChunk(id string, opts otf.GetChunkOptions) ([]byte, error) {
	var sql = "SELECT chunk, sequence FROM logs WHERE external_id = ?"
	var chunks [][]byte

	if err := db.DB.Select(&blob, sql, id); err != nil {
		return nil, databaseError(err)
	}

	return blob, nil
}
