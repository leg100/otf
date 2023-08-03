-- +goose Up
ALTER TABLE organizations ADD COLUMN cost_estimation_enabled BOOL DEFAULT false NOT NULL;
INSERT INTO run_statuses (status) VALUES ('cost_estimated');

-- +goose Down
DELETE FROM run_statuses WHERE status = 'cost_estimated';
ALTER TABLE organizations DROP COLUMN cost_estimation_enabled;
