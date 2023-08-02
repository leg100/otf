-- +goose Up
ALTER TABLE ingress_attributes
    ADD COLUMN commit_url TEXT,
    ADD COLUMN pull_request_number INTEGER,
    ADD COLUMN pull_request_url TEXT,
    ADD COLUMN pull_request_title TEXT,
    ADD COLUMN tag TEXT,
    ADD COLUMN sender_username TEXT,
    ADD COLUMN sender_avatar_url TEXT,
    ADD COLUMN sender_html_url TEXT;

UPDATE ingress_attributes
SET
    commit_url = '',
    pull_request_number = 0,
    pull_request_url = '',
    pull_request_title = '',
    tag = '',
    sender_username = '',
    sender_avatar_url = '',
    sender_html_url = '';

ALTER TABLE ingress_attributes
    ALTER COLUMN commit_url SET NOT NULL,
    ALTER COLUMN pull_request_number SET NOT NULL,
    ALTER COLUMN pull_request_url SET NOT NULL,
    ALTER COLUMN pull_request_title SET NOT NULL,
    ALTER COLUMN tag SET NOT NULL,
    ALTER COLUMN sender_username SET NOT NULL,
    ALTER COLUMN sender_avatar_url SET NOT NULL,
    ALTER COLUMN sender_html_url SET NOT NULL;

-- +goose Down
ALTER TABLE ingress_attributes
    DROP COLUMN commit_url,
    DROP COLUMN pull_request_number,
    DROP COLUMN pull_request_url,
    DROP COLUMN pull_request_title,
    DROP COLUMN tag,
    DROP COLUMN sender_username,
    DROP COLUMN sender_avatar_url,
    DROP COLUMN sender_html_url;
