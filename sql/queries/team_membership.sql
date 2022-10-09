-- name: InsertTeamMembership :exec
INSERT INTO team_memberships (
    user_id,
    team_id
) VALUES (
    pggen.arg('UserID'),
    pggen.arg('TeamID')
)
;

-- name: DeleteTeamMembership :one
DELETE
FROM team_memberships
WHERE
    user_id = pggen.arg('UserID') AND
    team_id = pggen.arg('TeamID')
RETURNING user_id
;
