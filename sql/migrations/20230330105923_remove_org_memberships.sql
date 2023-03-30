-- +goose Up
DROP TABLE IF EXISTS organization_memberships
;

-- +goose Down
CREATE TABLE IF NOT EXISTS organization_memberships (
    username          TEXT REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    organization_name TEXT REFERENCES organizations (name) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL
)
;
INSERT INTO organization_memberships (
    username,
    organization_name
)
SELECT DISTINCT u.username, t.organization_name
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
;
