package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

type chunk interface {
	GetChunk() []byte
	GetStart() *bool
	GetEnd() *bool
}

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

func mergeChunks(ctx context.Context, chunks []chunk, opts otf.GetChunkOptions) ([]byte, error) {
	// merge all chunks, prefixing or suffixing with start or end marker as
	// appropriate.
	var merged []byte
	for _, ch := range chunks {
		if *ch.GetStart() {
			merged = append(merged, otf.ChunkStartMarker)
		}

		merged = append(merged, ch.GetChunk()...)

		if *ch.GetEnd() {
			merged = append(merged, otf.ChunkEndMarker)
		}
	}

	return otf.GetChunk(merged, opts)
}
