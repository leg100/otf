-- Add job_id primary key to jobs table and populate with random identifiers.
ALTER TABLE jobs ADD COLUMN job_id TEXT;
UPDATE jobs SET job_id = 'job-' || substr(md5(random()::text), 0, 17);
ALTER TABLE jobs ADD PRIMARY KEY (job_id);

-- Add job_id primary key to jobs table and populate with random identifiers.
ALTER TABLE logs DROP chunk_id;
ALTER TABLE logs ADD chunk_id TEXT;
UPDATE logs SET chunk_id = 'chunk-' || substr(md5(random()::text), 0, 17);
ALTER TABLE logs ADD PRIMARY KEY (chunk_id);

---- create above / drop below ----

ALTER TABLE logs DROP chunk_id;
ALTER TABLE logs ADD chunk_id INT GENERATED ALWAYS AS IDENTITY;
ALTER TABLE logs ADD PRIMARY KEY (run_id, phase, chunk_id);

ALTER TABLE jobs DROP COLUMN job_id;
