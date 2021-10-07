package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/leg100/otf"
)

type CSV []string

func (c CSV) Value() (driver.Value, error) {
	return strings.Join(c, ","), nil
}

func (c CSV) Scan(src interface{}) error {
	c = strings.Split(src.(string), ",")
	return nil
}

type PlanTimeMap map[otf.PlanStatus]time.Time

func (m PlanTimeMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m PlanTimeMap) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), m)
}

type ApplyTimeMap map[otf.ApplyStatus]time.Time

func (m ApplyTimeMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m ApplyTimeMap) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), m)
}
