-- name: InsertTeam :exec
INSERT INTO teams (
    team_id,
    name,
    created_at,
    organization_id
) VALUES (
    pggen.arg('id'),
    pggen.arg('name'),
    pggen.arg('created_at'),
    pggen.arg('organization_id')
);

-- name: FindTeamByID :one
SELECT
    t.*,
    o.name AS organization_name
FROM teams t
JOIN organizations o USING (organization_id)
WHERE t.team_id = pggen.arg('team_id')
;

-- name: FindTeamByName :one
SELECT
    t.*,
    o.name AS organization_name
FROM teams t
JOIN organizations o USING (organization_id)
WHERE t.name = pggen.arg('name')
AND   o.name = pggen.arg('organization_name')
;

-- name: DeleteTeamByID :one
DELETE
FROM teams
WHERE team_id = pggen.arg('team_id')
RETURNING team_id
;

-- name: DeleteTeamByName :one
DELETE
FROM teams
USING organizations
WHERE teams.organization_id = organizations.organization_id
AND   teams.name = pggen.arg('name')
AND   organizations.name = pggen.arg('organization_name')
RETURNING team_id
;

