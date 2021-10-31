package sql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

func putChunk(ctx context.Context, db sqlx.Execer, table, idCol, idVal string, chunk []byte, opts otf.PutChunkOptions) error {
	insertBuilder := psql.Insert(table).
		Columns(idCol, "chunk", "start", "_end", "size").
		Values(idVal, chunk, opts.Start, opts.End, len(chunk))

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func getChunk(ctx context.Context, db Getter, table, idCol, idVal string, opts otf.GetChunkOptions) ([]byte, error) {
	type chunk struct {
		Data  []byte `db:"chunk"`
		Start bool
		End   bool `db:"_end"`
	}

	selectBuilder := psql.Select("chunk", "start", "_end").
		From(table).
		Where(fmt.Sprintf("%s = $1", idCol), idVal).
		OrderBy("chunk_id ASC")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var chunks []chunk
	if err := db.Select(&chunks, sql, args...); err != nil {
		return nil, databaseError(err)
	}

	// merge all chunks, prefixing or suffixing with start or end marker as
	// appropriate.
	var merged []byte
	for _, ch := range chunks {
		if ch.Start {
			merged = append(merged, otf.ChunkStartMarker)
		}

		merged = append(merged, ch.Data...)

		if ch.End {
			merged = append(merged, otf.ChunkEndMarker)
		}
	}

	if opts.Offset > len(merged) {
		return nil, fmt.Errorf("offset greater than size of logs: %d > %d", opts.Offset, len(merged))
	}

	if opts.Limit == 0 {
		return merged[opts.Offset:], nil
	}

	if opts.Limit > otf.ChunkMaxLimit {
		opts.Limit = otf.ChunkMaxLimit
	}

	// Adjust limit if it extends beyond size of logs
	if (opts.Offset + opts.Limit) > len(merged) {
		opts.Limit = len(merged) - opts.Offset
	}

	return merged[opts.Offset:(opts.Offset + opts.Limit)], nil
}
