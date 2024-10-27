-- Write your migrate up statements here
CREATE TABLE IF NOT EXISTS public.runners (
    runner_id text NOT NULL,
    version text NOT NULL,
    max_jobs integer NOT NULL,
    ip_address inet NOT NULL,
    last_ping_at timestamp with time zone NOT NULL,
    last_status_at timestamp with time zone NOT NULL,
    status text NOT NULL,
);

ALTER TABLE agents
  ALTER COLUMN 


---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
