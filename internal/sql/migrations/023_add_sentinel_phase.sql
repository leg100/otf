-- Add Sentinel as a first-class run phase.

CREATE TABLE IF NOT EXISTS sentinels (
    run_id text PRIMARY KEY,
    status text NOT NULL,
    CONSTRAINT sentinels_run_id_fkey FOREIGN KEY (run_id) REFERENCES runs(run_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT sentinels_status_fkey FOREIGN KEY (status) REFERENCES phase_statuses(status)
);

INSERT INTO phases (phase) VALUES ('sentinel') ON CONFLICT DO NOTHING;

INSERT INTO sentinels (run_id, status)
SELECT
    run_id,
    CASE
        WHEN status IN ('apply_queued', 'applying', 'applied', 'confirmed', 'policy_checked', 'policy_soft_failed', 'policy_failed') THEN 'finished'
        WHEN status IN ('pending', 'plan_queued', 'planning', 'planned', 'cost_estimated') THEN 'pending'
        ELSE 'unreachable'
    END
FROM runs
ON CONFLICT (run_id) DO NOTHING;
