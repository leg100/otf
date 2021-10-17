package otf

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type TimestampMap map[string]time.Time

func (m TimestampMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m TimestampMap) Scan(src interface{}) error {
	switch t := src.(type) {
	case string:
		return json.Unmarshal([]byte(t), &m)
	default:
		return fmt.Errorf("invalid type: %T", src)
	}
}
