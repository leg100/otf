package pubsub

import (
	"reflect"

	"github.com/leg100/otf/internal/sql"
)

type fakeListener struct{}

func (f *fakeListener) RegisterType(typ reflect.Type, ff sql.TableFunc) {
}
