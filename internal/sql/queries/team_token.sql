-- name: InsertTeamToken :exec
INSERT INTO team_tokens (
    team_token_id,
    description,
    created_at,
    team_id,
    expiry
) VALUES (
    pggen.arg('team_token_id'),
    pggen.arg('description'),
    pggen.arg('created_at'),
    pggen.arg('team_id'),
    pggen.arg('expiry')
);

--name: FindTeamTokensByTeam :many
SELECT *
FROM team_tokens
where team_id = pggen.arg('team_id')
;

--name: FindTeamTokenByID :one
SELECT *
FROM team_tokens
WHERE team_token_id = pggen.arg('team_token_id')
;

-- name: DeleteTeamTokenByID :one
DELETE
FROM team_tokens
WHERE team_token_id = pggen.arg('team_token_id')
RETURNING team_token_id
;