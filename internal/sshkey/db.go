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

func (db *pgdb) create(ctx context.Context, key *SSHKey, privateKey []byte) error {
	_, err := db.Exec(ctx, `
INSERT INTO ssh_keys (
    ssh_key_id,
    name,
    private_key,
    organization_name
) VALUES (
    @id,
    @name,
    @private_key,
    @organization_name
)
`,
		pgx.NamedArgs{
			"id":                key.ID,
			"name":              key.Name,
			"private_key":       privateKey,
			"organization_name": key.Organization,
		},
	)
	return err
}

func (db *pgdb) update(ctx context.Context, id resource.ID, updateFunc func(context.Context, *SSHKey) error) (*SSHKey, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*SSHKey, error) {
			rows := db.Query(ctx, `
SELECT ssh_key_id, name, organization_name
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
SET name = @name
WHERE ssh_key_id = @id
`,
				pgx.NamedArgs{
					"id":   key.ID,
					"name": key.Name,
				},
			)
			return err
		},
	)
}

func (db *pgdb) get(ctx context.Context, id resource.ID) (*SSHKey, error) {
	row := db.Query(ctx, `
SELECT ssh_key_id, name, organization_name
FROM ssh_keys
WHERE ssh_key_id = $1
`, id)
	return sql.CollectOneRow(row, pgx.RowToAddrOfStructByName[SSHKey])
}

func (db *pgdb) getPrivateKey(ctx context.Context, id resource.ID) ([]byte, error) {
	row := db.Query(ctx, `
SELECT private_key
FROM ssh_keys
WHERE ssh_key_id = $1
`, id)
	return sql.CollectOneType[[]byte](row)
}

func (db *pgdb) list(ctx context.Context, org organization.Name) ([]*SSHKey, error) {
	rows := db.Query(ctx, `
SELECT ssh_key_id, name, organization_name
FROM ssh_keys
WHERE organization_name = $1
ORDER BY name
`, org)
	return sql.CollectRows(rows, pgx.RowToAddrOfStructByName[SSHKey])
}

func (db *pgdb) delete(ctx context.Context, id resource.ID) (*SSHKey, error) {
	row := db.Query(ctx, `
DELETE FROM ssh_keys
WHERE ssh_key_id = $1
RETURNING ssh_key_id, name, organization_name
`, id)
	return sql.CollectOneRow(row, pgx.RowToAddrOfStructByName[SSHKey])
}
