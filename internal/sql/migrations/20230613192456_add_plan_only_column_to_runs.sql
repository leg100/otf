-- +goose Up
ALTER TABLE runs ADD COLUMN plan_only BOOL DEFAULT false;

UPDATE runs r
SET plan_only = cv.speculative
FROM configuration_versions cv
WHERE r.configuration_version_id = cv.configuration_version_id;

ALTER TABLE runs
    ALTER COLUMN plan_only SET NOT NULL,
    ALTER COLUMN plan_only DROP DEFAULT;

-- +goose Down
ALTER TABLE runs DROP COLUMN plan_only;
