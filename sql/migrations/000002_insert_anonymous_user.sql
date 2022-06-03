-- +goose Up
INSERT INTO users (
    user_id,
    username,
    created_at,
    updated_at
) VALUES (
    'user-anonymous',
    'anonymous',
    now(),
    now()
);

-- +goose Down
DELETE FROM users WHERE user_id = 'user-anonymous';
