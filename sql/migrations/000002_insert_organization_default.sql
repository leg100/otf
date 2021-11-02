-- +goose Up
INSERT INTO organizations (
    organization_id,
    created_at,
    updated_at,
    name,
    email,
    session_remember,
    session_timeout
) VALUES (
    'org-YYzEOzuSw4HjH8CW',
    now(),
    now(),
    'default',
    'admin@example.org',
    20160,
    20160
);

-- +goose Down
DELETE
FROM organizations
WHERE organization_id = 'org-YYzEOzuSw4HjH8CW';
