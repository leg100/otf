-- Update workspaces table to reference username instead of user ID for its user lock.

-- Add lock_username column and foreign key
ALTER TABLE workspaces ADD COLUMN lock_username TEXT;
ALTER TABLE workspaces ADD CONSTRAINT workspaces_lock_username_fkey
	FOREIGN KEY (lock_username) REFERENCES users(username)
	ON UPDATE CASCADE ON DELETE CASCADE;

-- Populate lock_username column
UPDATE workspaces w
SET lock_username = u.username
FROM users u
WHERE w.lock_user_id = u.user_id;

-- Drop lock_user_id column
ALTER TABLE workspaces DROP COLUMN lock_user_id;

---- create above / drop below ----

-- Add lock_user_id column and foreign key
ALTER TABLE workspaces ADD COLUMN lock_user_id TEXT;
ALTER TABLE workspaces ADD CONSTRAINT workspaces_lock_user_id_fkey
	FOREIGN KEY (lock_user_id) REFERENCES users(user_id)
	ON UPDATE CASCADE ON DELETE CASCADE;

-- Populate lock_user_id column
UPDATE workspaces w
SET lock_user_id = u.user_id
FROM users u
WHERE w.lock_username = u.username;

-- Drop lock_username column
ALTER TABLE workspaces DROP COLUMN lock_username;
