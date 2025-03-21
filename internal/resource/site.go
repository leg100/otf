package resource

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
)

// SiteID identifies the "site", a top-level abstraction identifying the OTF
// system as a whole.
var SiteID site

type site struct {
	// site should never be encoded/decoded or persisted to the db but must
	// satisfy the ID interface.
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	sql.Scanner
	driver.Valuer
}

func (s site) String() string { return "site" }
func (s site) Kind() Kind     { return SiteKind }
