-- name: FindQueuedJobs :many
SELECT
    j.job_id,
    j.run_id,
    j.status,
    r.is_destroy,
    r.refresh,
    r.refresh_only,
    w.auto_apply,
    cv.speculative,
    r.configuration_version_id,
    r.workspace_id
FROM jobs j
JOIN runs r ON r.run_id = j.run_id
JOIN configuration_versions cv USING(configuration_version_id)
JOIN workspaces w ON r.workspace_id = w.workspace_id
WHERE j.status = 'queued'
;
