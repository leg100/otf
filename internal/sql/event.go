package sql

import (
	"encoding/json"
	"log/slog"
	"time"
)

// Event is a postgres event triggered by a database change.
type Event struct {
	ID     int             `db:"id"`
	Table  Table           `db:"_table"` // pg table associated with change
	Action Action          // INSERT/UPDATE/DELETE
	Record json.RawMessage // the changed row
	Time   time.Time       // time at which event occured
}

func (v Event) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("action", string(v.Action)),
		slog.Any("table", v.Table),
		slog.Time("time", v.Time),
	}
	return slog.GroupValue(attrs...)
}
