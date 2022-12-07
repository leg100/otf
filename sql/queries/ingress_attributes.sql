-- name: InsertIngressAttributes :exec
INSERT INTO ingress_attributes (
    branch,
    commit_sha,
    identifier,
    is_pull_request,
    on_default_branch,
    configuration_version_id
) VALUES (
    pggen.arg('branch'),
    pggen.arg('commit_sha'),
    pggen.arg('identifier'),
    pggen.arg('is_pull_request'),
    pggen.arg('on_default_branch'),
    pggen.arg('configuration_version_id')
);
