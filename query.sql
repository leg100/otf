SELECT
    (
        SELECT (runs.*)::"runs"
        FROM runs
        JOIN workspaces ON w.lock_run_id = runs.run_id
    ) AS run_lock
FROM workspaces w
JOIN organizations o USING (organization_id)
WHERE o.name = 'automatize'
ORDER BY w.updated_at DESC
LIMIT 100
OFFSET 0
;
