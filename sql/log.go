package sql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

func putChunk(ctx context.Context, db sqlx.Execer, table, idCol, idVal string, chunk otf.Chunk) error {
	insertBuilder := psql.Insert(table).
		Columns(idCol, "chunk").
		Values(idVal, chunk.Marshal())

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

	return otf.UnmarshalChunk(data).Cut(opts)
}
