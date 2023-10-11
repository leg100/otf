-- name: UpdateLatestTerraformVersion :exec
UPDATE latest_terraform_version
SET version = pggen.arg('version');

-- name: UpdateLatestTerraformVersionCheckpoint :exec
UPDATE latest_terraform_version
SET checkpoint = current_timestamp;

-- name: FindLatestTerraformVersion :one
SELECT *
FROM latest_terraform_version;
