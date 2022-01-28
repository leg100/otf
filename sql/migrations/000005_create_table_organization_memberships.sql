-- +goose Up
CREATE TABLE IF NOT EXISTS organization_memberships (
    user_id text REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    organization_id text REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS organization_memberships;
