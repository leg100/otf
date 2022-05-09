package sql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

func putChunk(ctx context.Context, db sqlx.Execer, table, idCol, idVal string, chunk otf.Chunk) error {
	insertBuilder := psql.Insert(table).
		Column(idCol, "chunk").
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

func getChunk(ctx context.Context, db Getter, table, idCol, idVal string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	selectBuilder := psql.Select("string_agg(chunk, '')").
		From(table).
		Where(fmt.Sprintf("%s = $1", idCol), idVal).
		OrderBy("chunk_id ASC").

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return otf.Chunk{}, err
	}

	var chunks []otf.Chunk
	if err := db.Select(&chunks, sql, args...); err != nil {
		return otf.Chunk{}, databaseError(err, sql)
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

	return otf.GetChunk(merged, opts)
}
