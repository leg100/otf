-- name: InsertTeamMembership :exec
INSERT INTO team_memberships (
    username,
    team_id
) VALUES (
    pggen.arg('username'),
    pggen.arg('team_id')
)
;

-- name: DeleteTeamMembership :one
DELETE
FROM team_memberships
WHERE
    username = pggen.arg('username') AND
    team_id  = pggen.arg('team_id')
RETURNING username
;
