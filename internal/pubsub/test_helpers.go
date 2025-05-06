package pubsub

import "github.com/leg100/otf/internal/sql"

type fakeListener struct{}

func (f *fakeListener) RegisterFunc(table string, ff sql.TableFunc) {
}
