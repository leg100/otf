-- +goose Up
ALTER TABLE team_memberships ADD CONSTRAINT team_memberships_team_id_user_id_key UNIQUE (team_id, user_id);

-- +goose Down
ALTER TABLE team_memberships DROP CONSTRAINT team_memberships_team_id_user_id_key;
