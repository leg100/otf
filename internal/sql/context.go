package sql

import (
	"context"
)

const (
	// context key for retrieving connection from context
	connCtxKey ctxKey = 1
)

type ctxKey int

func newContext(ctx context.Context, conn genericConnection) context.Context {
	return context.WithValue(ctx, connCtxKey, conn)
}

func fromContext(ctx context.Context) (genericConnection, bool) {
	conn, ok := ctx.Value(connCtxKey).(genericConnection)
	return conn, ok
}
