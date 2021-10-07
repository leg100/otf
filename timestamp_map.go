package otf

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type TimestampMap map[string]time.Time

func (m TimestampMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m TimestampMap) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &m)
}
