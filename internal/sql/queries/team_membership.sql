-- name: InsertTeamMembership :many
WITH
    users AS (
        SELECT username
        FROM unnest(pggen.arg('usernames')::text[]) t(username)
    )
INSERT INTO team_memberships (username, team_id)
SELECT username, pggen.arg('team_id')
FROM users
RETURNING username
;

-- name: DeleteTeamMembership :many
WITH
    users AS (
        SELECT username
        FROM unnest(pggen.arg('usernames')::text[]) t(username)
    )
DELETE
FROM team_memberships tm
USING users
WHERE
    tm.username = users.username AND
    tm.team_id  = pggen.arg('team_id')
RETURNING tm.username
;
