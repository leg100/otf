package sql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

const (
	// ChunkMaxLimit is maximum permissible size of a chunk
	ChunkMaxLimit = 65536

	// ChunkStartMarker is the special byte that prefixes the first chunk
	ChunkStartMarker = byte(2)

	// ChunkEndMarker is the special byte that suffixes the last chunk
	ChunkEndMarker = byte(3)
)

var (
	_ otf.ChunkStore = (*LogDB)(nil)
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
func (db LogDB) PutChunk(id string, chunk []byte, opts otf.PutChunkOptions) error {
	sql := "INSERT INTO logs (log_id, chunk, start, _end, size) VALUES ($1, $2, $3, $4, $5)"

	_, err := db.Exec(sql, id, chunk, opts.Start, opts.End, len(chunk))
	if err != nil {
		return err
	}

	return nil
}

type chunk struct {
	Data  []byte `db:"chunk"`
	Start bool
	End   bool `db:"_end"`
}

func (db LogDB) GetChunk(id string, opts otf.GetChunkOptions) ([]byte, error) {
	var sql = "SELECT chunk, start, _end FROM logs WHERE log_id = $1 ORDER BY chunk_id ASC"
	var chunks []chunk

	if err := db.DB.Select(&chunks, sql, id); err != nil {
		return nil, databaseError(err)
	}

	// merge all chunks, prefixing or suffixing with start or end marker as
	// appropriate.
	var merged []byte
	for _, ch := range chunks {
		if ch.Start {
			merged = append(merged, ChunkStartMarker)
		}

		merged = append(merged, ch.Data...)

		if ch.End {
			merged = append(merged, ChunkEndMarker)
		}
	}

	if opts.Offset > len(merged) {
		return nil, fmt.Errorf("offset greater than size of logs: %d > %d", opts.Offset, len(merged))
	}

	if opts.Limit == 0 {
		return merged[opts.Offset:], nil
	}

	if opts.Limit > ChunkMaxLimit {
		opts.Limit = ChunkMaxLimit
	}

	// Adjust limit if it extends beyond size of logs
	if (opts.Offset + opts.Limit) > len(merged) {
		opts.Limit = len(merged) - opts.Offset
	}

	return merged[opts.Offset:(opts.Offset + opts.Limit)], nil
}
