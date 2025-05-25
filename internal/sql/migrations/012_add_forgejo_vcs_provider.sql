INSERT INTO vcs_kinds VALUES ('forgejo') ON CONFLICT DO NOTHING;
---- create above / drop below ----
DELETE FROM vcs_kinds WHERE name = 'forgejo';
