-- name: InsertVCSRepo :exec
INSERT INTO vcs_repos (
    identifier,
    branch,
    vcs_provider_id,
    workspace_id
) VALUES (
    pggen.arg('identifier'),
    pggen.arg('branch'),
    pggen.arg('vcs_provider_id'),
    pggen.arg('workspace_id')
);

-- name: UpdateVCSRepo :one
UPDATE vcs_repos
SET
    identifier = pggen.arg('identifier'),
    branch = pggen.arg('branch'),
    vcs_provider_id = pggen.arg('vcs_provider_id')
WHERE workspace_id = pggen.arg('workspace_id')
RETURNING workspace_id;

-- name: DeleteVCSRepo :exec
DELETE
FROM vcs_repos
WHERE workspace_id = pggen.arg('workspace_id')
RETURNING workspace_id
;
