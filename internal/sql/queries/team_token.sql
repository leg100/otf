-- name: InsertTeamToken :exec
INSERT INTO team_tokens (
    team_token_id,
    created_at,
    team_id,
    expiry
) VALUES (
    sqlc.arg('team_token_id'),
    sqlc.arg('created_at'),
    sqlc.arg('team_id'),
    sqlc.arg('expiry')
) ON CONFLICT (team_id) DO UPDATE
  SET team_token_id = sqlc.arg('team_token_id'),
      created_at    = sqlc.arg('created_at'),
      expiry        = sqlc.arg('expiry');

-- name: FindTeamTokensByID :many
SELECT *
FROM team_tokens
WHERE team_id = sqlc.arg('team_id')
;

-- name: DeleteTeamTokenByID :one
DELETE
FROM team_tokens
WHERE team_id = sqlc.arg('team_id')
RETURNING team_token_id
;
