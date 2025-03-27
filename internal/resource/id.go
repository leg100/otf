package resource

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"regexp"
)

type ID interface {
	fmt.Stringer
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	sql.Scanner
	driver.Valuer

	Kind() Kind
}

// ReStringID is a regular expression used to validate common string ID patterns.
var ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
