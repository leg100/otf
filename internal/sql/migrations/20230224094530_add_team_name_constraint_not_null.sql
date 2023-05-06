-- +goose Up
ALTER TABLE teams
    ALTER COLUMN name SET NOT NULL,
    ADD CONSTRAINT team_name_uniq UNIQUE(organization_name, name)
;

-- +goose Down
ALTER TABLE teams
    ALTER COLUMN name DROP NOT NULL,
    DROP CONSTRAINT team_name_uniq
;
