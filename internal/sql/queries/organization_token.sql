-- name: UpsertOrganizationToken :exec
INSERT INTO organization_tokens (
    organization_token_id,
    created_at,
    organization_name,
    expiry
) VALUES (
    pggen.arg('organization_token_id'),
    pggen.arg('created_at'),
    pggen.arg('organization_name'),
    pggen.arg('expiry')
) ON CONFLICT (organization_name) DO UPDATE
  SET created_at            = pggen.arg('created_at'),
      organization_token_id = pggen.arg('organization_token_id'),
      expiry                = pggen.arg('expiry');

-- name: FindOrganizationTokensByName :many
SELECT *
FROM organization_tokens
WHERE organization_name = pggen.arg('organization_name');

-- name: FindOrganizationTokensByID :one
SELECT *
FROM organization_tokens
WHERE organization_token_id = pggen.arg('organization_token_id');

-- name: DeleteOrganiationTokenByName :one
DELETE
FROM organization_tokens
WHERE organization_name = pggen.arg('organization_name')
RETURNING organization_token_id;
