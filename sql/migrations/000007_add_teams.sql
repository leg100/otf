-- +goose Up
CREATE TABLE IF NOT EXISTS teams (
    team_id           TEXT,
    name              TEXT,
    created_at        TIMESTAMPTZ NOT NULL,
    organization_id   TEXT REFERENCES organizations ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
                      PRIMARY KEY (team_id),
                      UNIQUE (name, organization_id)
);

CREATE TABLE IF NOT EXISTS team_memberships (
    team_id         TEXT REFERENCES teams ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    user_id         TEXT REFERENCES users ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS team_memberships;
DROP TABLE IF EXISTS teams;
