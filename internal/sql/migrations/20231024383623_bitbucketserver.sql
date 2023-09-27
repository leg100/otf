-- +goose Up

INSERT INTO vcs_kinds (name) VALUES
	('bitbucketserver');

-- +goose Down

DELETE FROM vcs_kinds WHERE name = 'bitbucketserver';