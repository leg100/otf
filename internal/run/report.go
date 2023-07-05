package run

import (
	"fmt"

	"github.com/leg100/otf/internal/sql/pggen"
)

// Report reports a summary of additions, changes, and deletions of
// resources in a plan or an apply.
type Report struct {
	Additions    int `json:"additions"`
	Changes      int `json:"changes"`
	Destructions int `json:"destructions"`
}

func reportFromDB(row *pggen.Report) *Report {
	if row == nil {
		return nil
	}
	return &Report{
		Additions:    int(row.Additions.Int),
		Changes:      int(row.Changes.Int),
		Destructions: int(row.Destructions.Int),
	}
}

func (r Report) HasChanges() bool {
	return r != Report{}
}

func (r Report) String() string {
	// \u2212 is a proper minus sign; an ascii hyphen is too narrow (in the
	// default github font at least) and looks incongruous alongside
	// the wider '+' and '~' characters.
	return fmt.Sprintf("+%d/~%d/\u2212%d", r.Additions, r.Changes, r.Destructions)
}
