CREATE INDEX IF NOT EXISTS idx_logs_run_id ON logs(run_id);
---- create above / drop below ----
DROP INDEX idx_logs_run_id;
