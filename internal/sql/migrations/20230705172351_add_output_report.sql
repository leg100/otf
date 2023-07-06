-- +goose Up
ALTER TABLE plans
    RENAME COLUMN report TO resource_report;
ALTER TABLE applies
    RENAME COLUMN report TO resource_report;

ALTER TABLE plans
    ADD COLUMN output_report REPORT;

-- +goose Down
ALTER TABLE plans
    RENAME COLUMN resource_report TO report;
ALTER TABLE applies
    RENAME COLUMN resource_report TO report;

ALTER TABLE plans
    DROP COLUMN output_report;
