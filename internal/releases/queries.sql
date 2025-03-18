-- name: InsertLatestTerraformVersion :exec
INSERT INTO latest_terraform_version (
    version,
    checkpoint
) VALUES (
    sqlc.arg('version'),
    current_timestamp
);

-- name: UpdateLatestTerraformVersion :exec
UPDATE latest_terraform_version
SET version = sqlc.arg('version'),
    checkpoint = current_timestamp;

-- name: FindLatestTerraformVersion :many
SELECT *
FROM latest_terraform_version;
