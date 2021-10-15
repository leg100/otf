package otf

import (
	"database/sql/driver"
	"strings"
)

type CSV []string

func (c CSV) Value() (driver.Value, error) {
	return strings.Join(c, ","), nil
}

func (c *CSV) Scan(src interface{}) error {
	if src.(string) == "" {
		return nil
	}

	*c = strings.Split(src.(string), ",")

	return nil
}
