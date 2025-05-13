package sql

import (
	"context"
)

const (
	// context key for retrieving connection from context
	connCtxKey ctxKey = 1
)

type ctxKey int

func newContext(ctx context.Context, conn connection) context.Context {
	return context.WithValue(ctx, connCtxKey, conn)
}

func fromContext(ctx context.Context) (connection, bool) {
	conn, ok := ctx.Value(connCtxKey).(connection)
	return conn, ok
}
