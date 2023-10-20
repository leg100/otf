-- +goose Up
CREATE TABLE IF NOT EXISTS team_tokens (
    team_token_id   TEXT NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL,
    team_id         TEXT REFERENCES teams (team_id) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    expiry          TIMESTAMPTZ,
                    PRIMARY KEY (team_token_id),
                    UNIQUE (team_token_id)
);

-- +goose Down
DROP TABLE IF EXISTS team_tokens;