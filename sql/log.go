package sql

import (
	"context"
	"fmt"

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

func getChunk(ctx context.Context, db Getter, table, idCol, idVal string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	selectBuilder := psql.
		Select("string_agg(chunk, '')").
		FromSelect(
			psql.Select(idCol, "chunk").
				From(table).
				Where(fmt.Sprintf("%s = ?", idCol), idVal).
				OrderBy("chunk_id ASC"),
			"t").
		GroupBy(idCol)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return otf.Chunk{}, err
	}

	var data []byte
	if err := db.Get(&data, sql, args...); err != nil {
		return otf.Chunk{}, databaseError(err, sql)
	}

	return otf.UnmarshalChunk(data).Cut(opts)
}
