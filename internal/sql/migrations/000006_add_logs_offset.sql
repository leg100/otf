-- +goose Up
ALTER TABLE logs ADD COLUMN _offset INTEGER NOT NULL DEFAULT 0;
-- +goose StatementBegin
CREATE FUNCTION populate_offsets() RETURNS void AS $$
DECLARE
    chunks RECORD;
    emptystring BYTEA := '';
BEGIN
    FOR chunks IN
        SELECT chunk_id, run_id, phase
        FROM logs
    LOOP
        EXECUTE 'WITH offsets AS ('
                || 'SELECT length(string_agg(chunk, $4)) AS offset '
                || 'FROM logs '
                || 'WHERE chunk_id < $1 '
                || 'AND run_id = $2 '
                || 'AND phase = $3'
            || ') UPDATE logs SET _offset = offsets.offset FROM offsets '
            || 'WHERE chunk_id = $1 AND offsets.offset IS NOT NULL'
            USING chunks.chunk_id, chunks.run_id, chunks.phase, emptystring;
    END LOOP;
    RAISE NOTICE 'Done populating log offsets';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
SELECT populate_offsets();
DROP FUNCTION populate_offsets;

-- +goose Down
ALTER TABLE logs DROP COLUMN _offset;
