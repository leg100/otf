-- name: InsertTeamToken :exec
INSERT INTO team_tokens (
    team_token_id,
    created_at,
    team_id,
    expiry
) VALUES (
    pggen.arg('team_token_id'),
    pggen.arg('created_at'),
    pggen.arg('team_id'),
    pggen.arg('expiry')
) ON CONFLICT (team_id) DO UPDATE
  SET team_token_id = pggen.arg('team_token_id'),
      created_at    = pggen.arg('created_at'),
      expiry        = pggen.arg('expiry');

--name: FindTeamTokensByID :many
SELECT *
FROM team_tokens
WHERE team_id = pggen.arg('team_id')
;

-- name: DeleteTeamTokenByID :one
DELETE
FROM team_tokens
WHERE team_id = pggen.arg('team_id')
RETURNING team_token_id
;
