package pubsub

import (
	"github.com/leg100/otf/internal/sql"
)

type fakeListener struct{}

func (f *fakeListener) RegisterTable(table string, ff sql.ForwardFunc) {
}
