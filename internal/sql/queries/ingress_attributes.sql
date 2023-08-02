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
    pggen.arg('branch'),
    pggen.arg('commit_sha'),
    pggen.arg('commit_url'),
    pggen.arg('pull_request_number'),
    pggen.arg('pull_request_url'),
    pggen.arg('pull_request_title'),
    pggen.arg('sender_username'),
    pggen.arg('sender_avatar_url'),
    pggen.arg('sender_html_url'),
    pggen.arg('identifier'),
    pggen.arg('tag'),
    pggen.arg('is_pull_request'),
    pggen.arg('on_default_branch'),
    pggen.arg('configuration_version_id')
);
