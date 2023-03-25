package pubsub

import (
	"context"

	"github.com/jackc/pgconn"
)

type fakePool struct {
	gotExecArgs []any

	pool
}

func (f *fakePool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	f.gotExecArgs = arguments
	return nil, nil
}

type fakeGetter struct {
	resource any
}

func (f *fakeGetter) GetByID(context.Context, string) (any, error) {
	return f.resource, nil
}
