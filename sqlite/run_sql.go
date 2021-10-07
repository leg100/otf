package sqlite

import "text/template"

var insertRunSql = `INSERT INTO runs (created_at, updated_at, external_id, force_cancel_available_at, is_destroy, message, position_in_queue, refresh, refresh_only, status, status_timestamps, replace_addrs, target_addrs, workspace_id, configuration_version_id)
VALUES (:created_at,:updated_at,:external_id,:force_cancel_available_at,?,?,?,?,?,?,?,?,?,?)
`

var insertPlanSql = `INSERT INTO plans (created_at, updated_at, external_id, logs_blob_id, plan_file_blob_id, plan_json_blob_id, status, status_timestamps, run_id
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
`

var insertApplySql = `INSERT INTO applies (created_at, updated_at, external_id, logs_blob_id, status, status_timestamps, run_id
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
`

var updateRunSql = template.Must(template.New("updateRunSql").Parse(`UPDATE runs
{{range .Fields}}
SET {{.}} = :{{.}}
WHERE id = :id
{{end}}
`))

var getRunColumns = `runs.*,
plans.id AS plans.id,
plans.created_at AS plans.created_at,
plans.updated_at AS plans.updated_at,
plans.resource_additions AS plans.resource_additions,
plans.resource_changes AS plans.resource_changes,
plans.resource_deletions AS plans.resource_deletions,
plans.status AS plans.status,
plans.status_timestamps AS plans.status_timestamps,
plans.logs_blob_id AS plans.logs_blob_id,
plans.plan_file_blob_id AS plans.plan_file_blob_id,
plans.plan_json_blob_id AS plans.plan_json_blob_id,
plans.run_id AS plans.run_id,
applies.id AS applies.id,
applies.created_at AS applies.created_at,
applies.updated_at AS applies.updated_at,
applies.resource_additions AS applies.resource_additions,
applies.resource_changes AS applies.resource_changes,
applies.resource_deletions AS applies.resource_deletions,
applies.status AS applies.status,
applies.status_timestamps AS applies.status_timestamps,
applies.logs_blob_id AS applies.logs_blob_id,
applies.run_id AS applies.run_id
workspaces.status AS workspaces.status
workspaces.external_id AS workspaces.external_id
`

var getRunJoins = `
JOIN plans ON plans.run_id = runs.id
JOIN applies ON applies.run_id = runs.id
JOIN configuration_versions ON configuration_versions.id = runs.configuration_version_id
JOIN workspaces ON workspaces.id = runs.workspace_id
`

var getRunSql = template.Must(template.New("getRunSql").Parse(`SELECT
runs.*,
plans.id AS plans.id,
plans.created_at AS plans.created_at,
plans.updated_at AS plans.updated_at,
plans.resource_additions AS plans.resource_additions,
plans.resource_changes AS plans.resource_changes,
plans.resource_deletions AS plans.resource_deletions,
plans.status AS plans.status,
plans.status_timestamps AS plans.status_timestamps,
plans.logs_blob_id AS plans.logs_blob_id,
plans.plan_file_blob_id AS plans.plan_file_blob_id,
plans.plan_json_blob_id AS plans.plan_json_blob_id,
plans.run_id AS plans.run_id,
applies.id AS applies.id,
applies.created_at AS applies.created_at,
applies.updated_at AS applies.updated_at,
applies.resource_additions AS applies.resource_additions,
applies.resource_changes AS applies.resource_changes,
applies.resource_deletions AS applies.resource_deletions,
applies.status AS applies.status,
applies.status_timestamps AS applies.status_timestamps,
applies.logs_blob_id AS applies.logs_blob_id,
applies.run_id AS applies.run_id
workspaces.status AS workspaces.status
workspaces.external_id AS workspaces.external_id
FROM runs
JOIN plans ON plans.run_id = runs.id
JOIN applies ON applies.run_id = runs.id
JOIN configuration_versions ON configuration_versions.id = runs.configuration_version_id
JOIN workspaces ON workspaces.id = runs.workspace_id
WHERE
{{range $key, $value := .Predicates }}
{{if .WorkspaceID }} runs.workspace_id = :workspace_id {{end}}
{{if .ID }} runs.id = :id {{end}}
{{if .PlanID }} plans.id = :plan_id {{end}}
{{if .ApplyID }} applies.id = :apply_id {{end}}
{{if .Statuses }} AND status IN (:statuses) {{end}}
{{end}}
{{if .Limit }} LIMIT :limit {{end}}
{{if .Offset }} OFFSET :offset {{end}}
`))
