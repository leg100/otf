ALTER TABLE jobs DROP COLUMN signaled;
---- create above / drop below ----
ALTER TABLE jobs ADD COLUMN signaled BOOLEAN;
