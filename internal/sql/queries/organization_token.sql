-- name: UpsertOrganizationToken :exec
INSERT INTO organization_tokens (
    organization_token_id,
    created_at,
    organization_name,
    expiry
) VALUES (
    sqlc.arg('organization_token_id'),
    sqlc.arg('created_at'),
    sqlc.arg('organization_name'),
    sqlc.arg('expiry')
) ON CONFLICT (organization_name) DO UPDATE
  SET created_at            = sqlc.arg('created_at'),
      organization_token_id = sqlc.arg('organization_token_id'),
      expiry                = sqlc.arg('expiry');

-- name: FindOrganizationTokens :many
SELECT *
FROM organization_tokens
WHERE organization_name = sqlc.arg('organization_name');

-- name: FindOrganizationTokensByName :one
SELECT *
FROM organization_tokens
WHERE organization_name = sqlc.arg('organization_name');

-- name: FindOrganizationTokensByID :one
SELECT *
FROM organization_tokens
WHERE organization_token_id = sqlc.arg('organization_token_id');

-- name: DeleteOrganiationTokenByName :one
DELETE
FROM organization_tokens
WHERE organization_name = sqlc.arg('organization_name')
RETURNING organization_token_id;
