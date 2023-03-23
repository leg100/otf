package pubsub

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/jackc/pgconn"
	"github.com/leg100/otf"
)

func NewTestBroker(t *testing.T, db otf.DB) Broker {
	return NewBroker(logr.Discard(), db)
}

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
