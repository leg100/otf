ALTER TABLE users ADD COLUMN avatar BYTEA;
---- create above / drop below ----
ALTER TABLE users DROP COLUMN avatar;
