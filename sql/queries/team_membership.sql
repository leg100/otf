-- name: InsertTeamMembership :exec
INSERT INTO team_memberships (
    user_id,
    team_id
) VALUES (
    pggen.arg('user_id'),
    pggen.arg('team_id')
)
;

-- name: DeleteTeamMembership :one
DELETE
FROM team_memberships
WHERE
    user_id = pggen.arg('user_id') AND
    team_id = pggen.arg('team_id')
RETURNING user_id
;
