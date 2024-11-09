-- Add job_id primary key to jobs table and populate with random identifiers.
ALTER TABLE jobs ADD COLUMN job_id TEXT;
UPDATE jobs SET job_id = 'job-' || substr(md5(random()::text), 0, 17);
ALTER TABLE jobs ADD PRIMARY KEY (job_id);
-- Add unique constraint - there should only be one job per run/phase combo
ALTER TABLE jobs ADD UNIQUE (run_id, phase);

-- replace job event function to instead provide job_id in payload
CREATE OR REPLACE FUNCTION public.jobs_notify_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    record RECORD;
    notification JSON;
BEGIN
    IF (TG_OP = 'DELETE') THEN
        record = OLD;
    ELSE
        record = NEW;
    END IF;
    notification = json_build_object(
                      'table',TG_TABLE_NAME,
                      'action', TG_OP,
                      'id', record.job_id);
    PERFORM pg_notify('events', notification::text);
    RETURN NULL;
END;
$$;

-- Change logs's chunk_id from a generated serial type to a text type and populate with random resource IDs
ALTER TABLE logs DROP chunk_id;
ALTER TABLE logs ADD chunk_id TEXT;
UPDATE logs SET chunk_id = 'chunk-' || substr(md5(random()::text), 0, 17);
ALTER TABLE logs ADD PRIMARY KEY (chunk_id);

-- Change workspace's lock_username column to lock_user_id
ALTER TABLE workspaces DROP COLUMN lock_username;
ALTER TABLE workspaces ADD COLUMN lock_user_id TEXT;
ALTER TABLE workspaces ADD FOREIGN KEY (lock_user_id) REFERENCES users(user_id);

-- Make site-admin's user ID valid with new stricter ID format
UPDATE users SET user_id = 'user-36atQC2oGQng7pVz' WHERE username = 'site-admin';

---- create above / drop below ----

UPDATE users SET user_id = 'user-site-admin' WHERE username = 'site-admin';

ALTER TABLE workspaces DROP COLUMN lock_user_id;
ALTER TABLE workspaces ADD COLUMN lock_username TEXT;
ALTER TABLE workspaces ADD FOREIGN KEY (lock_username) REFERENCES users(username);

ALTER TABLE logs DROP chunk_id;
ALTER TABLE logs ADD chunk_id INT GENERATED ALWAYS AS IDENTITY;
ALTER TABLE logs ADD PRIMARY KEY (run_id, phase, chunk_id);

ALTER TABLE jobs DROP CONSTRAINT jobs_run_id_phase_key;
ALTER TABLE jobs DROP COLUMN job_id;
