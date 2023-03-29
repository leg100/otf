--
-- migrate foreign keys from user id to username
--
-- +goose Up
--
-- migrate fk on organization memberships
--
ALTER TABLE organization_memberships
    ADD COLUMN username TEXT,
    ADD CONSTRAINT org_member_username_fk
        FOREIGN KEY (username) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE organization_memberships om
SET username = u.username
FROM users u
WHERE om.user_id = u.user_id
;
ALTER TABLE organization_memberships
    ALTER COLUMN username SET NOT NULL,
    DROP COLUMN user_id
;
--
-- migrate fk on team memberships
--
ALTER TABLE team_memberships
    ADD COLUMN username TEXT,
    ADD CONSTRAINT team_member_username_fk
        FOREIGN KEY (username) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE team_memberships om
SET username = u.username
FROM users u
WHERE om.user_id = u.user_id
;
ALTER TABLE team_memberships
    ALTER COLUMN username SET NOT NULL,
    DROP COLUMN user_id
;
--
-- migrate fk on workspaces (the user lock)
--
ALTER TABLE workspaces
    ADD COLUMN lock_username TEXT,
    ADD CONSTRAINT workspace_lock_username_fk
        FOREIGN KEY (lock_username) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE workspaces w
SET lock_username = u.username
FROM users u
WHERE w.lock_user_id = u.user_id
;
ALTER TABLE workspaces
    DROP COLUMN lock_user_id
;
--
-- migrate fk on sessions
--
ALTER TABLE sessions
    ADD COLUMN username TEXT,
    ADD CONSTRAINT session_username_fk
        FOREIGN KEY (username) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE sessions s
SET username = u.username
FROM users u
WHERE s.user_id = u.user_id
;
ALTER TABLE sessions
    ALTER COLUMN username SET NOT NULL,
    DROP COLUMN user_id
;
--
-- migrate fk on tokens
--
ALTER TABLE tokens
    ADD COLUMN username TEXT,
    ADD CONSTRAINT token_username_fk
        FOREIGN KEY (username) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE tokens t
SET username = u.username
FROM users u
WHERE t.user_id = u.user_id
;
ALTER TABLE tokens
    ALTER COLUMN username SET NOT NULL,
    DROP COLUMN user_id
;

-- +goose Down
--
-- migrate fk on organization memberships
--
ALTER TABLE organization_memberships
    ADD COLUMN user_id TEXT,
    ADD CONSTRAINT org_member_userid_fk
        FOREIGN KEY (user_id) REFERENCES users (user_id) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE organization_memberships om
SET user_id = u.user_id
FROM users u
WHERE om.username = u.username
;
ALTER TABLE organization_memberships
    ALTER COLUMN user_id SET NOT NULL,
    DROP COLUMN username
;
--
-- migrate fk on team memberships
--
ALTER TABLE team_memberships
    ADD COLUMN user_id TEXT,
    ADD CONSTRAINT team_member_userid_fk
        FOREIGN KEY (user_id) REFERENCES users (user_id) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE team_memberships om
SET user_id = u.user_id
FROM users u
WHERE om.username = u.username
;
ALTER TABLE team_memberships
    ALTER COLUMN user_id SET NOT NULL,
    DROP COLUMN username
;
--
-- migrate fk on workspaces (user lock)
--
ALTER TABLE workspaces
    ADD COLUMN lock_user_id TEXT,
    ADD CONSTRAINT workspace_lock_user_id_fk
        FOREIGN KEY (lock_user_id) REFERENCES users (user_id) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE workspaces w
SET lock_user_id = u.user_id
FROM users u
WHERE w.lock_username = u.username
;
ALTER TABLE workspaces
    DROP COLUMN lock_username
;
--
-- migrate fk on sessions
--
ALTER TABLE sessions
    ADD COLUMN user_id TEXT,
    ADD CONSTRAINT session_user_id_fk
        FOREIGN KEY (user_id) REFERENCES users (user_id) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE sessions s
SET user_id = u.user_id
FROM users u
WHERE s.username = u.username
;
ALTER TABLE sessions
    ALTER COLUMN user_id SET NOT NULL,
    DROP COLUMN username
;
--
-- migrate fk on tokens
--
ALTER TABLE tokens
    ADD COLUMN user_id TEXT,
    ADD CONSTRAINT token_user_id_fk
        FOREIGN KEY (user_id) REFERENCES users (user_id) ON UPDATE CASCADE ON DELETE CASCADE
;
UPDATE tokens t
SET user_id = u.user_id
FROM users u
WHERE t.username = u.username
;
ALTER TABLE tokens
    ALTER COLUMN user_id SET NOT NULL,
    DROP COLUMN username
;
