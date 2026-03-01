package sshkey

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type pgdb struct {
	*sql.DB
}

func (db *pgdb) create(ctx context.Context, key *SSHKey) error {
	_, err := db.Exec(ctx, `
INSERT INTO ssh_keys (
    ssh_key_id,
    created_at,
    updated_at,
    name,
    private_key,
    organization_name
) VALUES (
    @id,
    @created_at,
    @updated_at,
    @name,
    @private_key,
    @organization_name
)
`,
		pgx.NamedArgs{
			"id":                key.ID,
			"created_at":        key.CreatedAt,
			"updated_at":        key.UpdatedAt,
			"name":              key.Name,
			"private_key":       key.PrivateKey,
			"organization_name": key.Organization,
		},
	)
	return err
}

func (db *pgdb) update(ctx context.Context, id resource.TfeID, updateFunc func(context.Context, *SSHKey) error) (*SSHKey, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*SSHKey, error) {
			rows := db.Query(ctx, `
SELECT *
FROM ssh_keys
WHERE ssh_key_id = $1
FOR UPDATE
`, id)
			return sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[SSHKey])
		},
		updateFunc,
		func(ctx context.Context, key *SSHKey) error {
			_, err := db.Exec(ctx, `
UPDATE ssh_keys
SET
    updated_at  = @updated_at,
    name        = @name,
    private_key = @private_key
WHERE ssh_key_id = @id
`,
				pgx.NamedArgs{
					"id":         key.ID,
					"updated_at": key.UpdatedAt,
					"name":       key.Name,
					"private_key": key.PrivateKey,
				},
			)
			return err
		},
	)
}

func (db *pgdb) get(ctx context.Context, id resource.TfeID) (*SSHKey, error) {
	rows := db.Query(ctx, `
SELECT *
FROM ssh_keys
WHERE ssh_key_id = $1
`, id)
	return sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[SSHKey])
}

func (db *pgdb) list(ctx context.Context, org organization.Name) ([]*SSHKey, error) {
	rows := db.Query(ctx, `
SELECT *
FROM ssh_keys
WHERE organization_name = $1
ORDER BY name
`, org)
	return sql.CollectRows(rows, pgx.RowToAddrOfStructByName[SSHKey])
}

func (db *pgdb) delete(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE FROM ssh_keys
WHERE ssh_key_id = $1
`, id)
	return err
}
