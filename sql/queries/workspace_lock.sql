-- name: InsertWorkspaceLockUser :exec
INSERT INTO workspace_locks (
    workspace_id,
    user_id
) VALUES (
    pggen.arg('workspace_id'),
    pggen.arg('user_id')
);

-- name: InsertWorkspaceLockRun :exec
INSERT INTO workspace_locks (
    workspace_id,
    run_id
) VALUES (
    pggen.arg('workspace_id'),
    pggen.arg('run_id')
);

-- name: FindWorkspaceLockForUpdate :one
SELECT l.*
FROM workspace_locks l
LEFT JOIN users USING (user_id)
LEFT JOIN runs USING (run_id)
WHERE l.workspace_id = pggen.arg('workspace_id')
FOR UPDATE of l;

-- name: DeleteWorkspaceLock :one
DELETE
FROM workspace_locks
WHERE workspace_id = pggen.arg('workspace_id')
RETURNING workspace_id;
