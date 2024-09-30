-- name: InsertIngressAttributes :exec
INSERT INTO ingress_attributes (
    branch,
    commit_sha,
    commit_url,
    pull_request_number,
    pull_request_url,
    pull_request_title,
    sender_username,
    sender_avatar_url,
    sender_html_url,
    identifier,
    tag,
    is_pull_request,
    on_default_branch,
    configuration_version_id
) VALUES (
    sqlc.arg('branch'),
    sqlc.arg('commit_sha'),
    sqlc.arg('commit_url'),
    sqlc.arg('pull_request_number'),
    sqlc.arg('pull_request_url'),
    sqlc.arg('pull_request_title'),
    sqlc.arg('sender_username'),
    sqlc.arg('sender_avatar_url'),
    sqlc.arg('sender_html_url'),
    sqlc.arg('identifier'),
    sqlc.arg('tag'),
    sqlc.arg('is_pull_request'),
    sqlc.arg('on_default_branch'),
    sqlc.arg('configuration_version_id')
);
