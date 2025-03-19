-- name: InsertRepohook :one
WITH inserted AS (
    INSERT INTO repohooks (
        repohook_id,
        vcs_id,
        vcs_provider_id,
        secret,
        repo_path
    ) VALUES (
        sqlc.arg('repohook_id'),
        sqlc.arg('vcs_id'),
        sqlc.arg('vcs_provider_id'),
        sqlc.arg('secret'),
        sqlc.arg('repo_path')
    )
    RETURNING *
)
SELECT
    w.repohook_id,
    w.vcs_id,
    v.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM inserted w
JOIN vcs_providers v USING (vcs_provider_id);

-- name: UpdateRepohookVCSID :one
UPDATE repohooks
SET vcs_id = sqlc.arg('vcs_id')
WHERE repohook_id = sqlc.arg('repohook_id')
RETURNING *;

-- name: FindRepohooks :many
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id);

-- name: FindRepohookByID :one
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE w.repohook_id = sqlc.arg('repohook_id');

-- name: FindRepohookByRepoAndProvider :many
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE repo_path = sqlc.arg('repo_path')
AND   w.vcs_provider_id = sqlc.arg('vcs_provider_id');

-- name: FindUnreferencedRepohooks :many
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE NOT EXISTS (
    SELECT FROM repo_connections rc
    WHERE rc.vcs_provider_id = w.vcs_provider_id
    AND   rc.repo_path = w.repo_path
);

-- name: DeleteRepohookByID :one
DELETE
FROM repohooks
WHERE repohook_id = sqlc.arg('repohook_id')
RETURNING *;
