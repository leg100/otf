package resource

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"regexp"
)

// ReStringID is a regular expression used to validate common string ID patterns.
var ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)

// ID uniquely identifies an OTF resource.
type ID interface {
	fmt.Stringer
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	sql.Scanner
	driver.Valuer
	Kind() Kind
}
