-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION delete_tags() RETURNS TRIGGER AS $$
BEGIN
    DELETE
    FROM tags
    WHERE NOT EXISTS (
        SELECT FROM workspace_tags wt
        WHERE wt.tag_id = tags.tag_id
    );
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER delete_tags
AFTER DELETE ON workspace_tags
    FOR EACH STATEMENT EXECUTE FUNCTION delete_tags();

-- +goose Down
DROP TRIGGER IF EXISTS delete_tags ON workspace_tags;
DROP FUNCTION IF EXISTS delete_tags;
