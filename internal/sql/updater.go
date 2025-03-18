package sql

import (
	"context"
)

// Updater handles the common flow of a database update:
// 1) open tx
// 2) retrieve row
// 3) convert row to type
// 4) invoke update function on type
// 5) update database with updated type
// 6) close tx
//
// If an error occurs the tx is rolled back.
//
// The updater also ensures the same correct context is passed to each function.
// This is especially important because we pass the tx within the context to
// permit the update function to re-use the tx; using the wrong context would
// mean the update function does not re-use the tx, and instead create another
// tx, which not only means the outer update function is no longer atomic, but
// it creates the possibility for deadlock because it would attempt to acquire *two*
// database connections, and it'll do so indefinitely if the max
// connection count has been hit and there are other update calls in-flight also
// waiting for two connections.
func Updater[T any](
	ctx context.Context,
	db *DB,
	getForUpdate func(context.Context, Connection) (T, error),
	update func(context.Context, T) error,
	updateDB func(context.Context, Connection, T) error,
) (T, error) {
	var row T
	err := db.Tx(ctx, func(ctx context.Context, conn Connection) error {
		var err error
		row, err = getForUpdate(ctx, conn)
		if err != nil {
			return err
		}
		if err := update(ctx, row); err != nil {
			return err
		}
		if err := updateDB(ctx, conn, row); err != nil {
			return err
		}
		return nil
	})
	return row, Error(err)
}
