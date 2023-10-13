-- name: InsertLatestTerraformVersion :exec
INSERT INTO latest_terraform_version (
    version,
    checkpoint
) VALUES (
    pggen.arg('version'),
    current_timestamp
);

-- name: UpdateLatestTerraformVersion :exec
UPDATE latest_terraform_version
SET version = pggen.arg('version'),
    checkpoint = current_timestamp;

-- name: FindLatestTerraformVersion :many
SELECT *
FROM latest_terraform_version;
