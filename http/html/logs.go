package html

import (
	"html/template"
	"strings"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/leg100/otf"
)

// logs represents existing logs retreived from the db for a run phase
type logs otf.Chunk

// Existing returns the existing logs in HTML format.
func (l *logs) Existing() template.HTML {
	// convert to string
	logs := string(l.Data)
	// convert ANSI escape sequences to HTML
	logs = string(term2html.Render([]byte(logs)))
	// trim leading and trailing white space
	logs = strings.TrimSpace(logs)

	return template.HTML(logs)
}

// Offset returns the offset from which logs can be tailed
func (l *logs) Offset() int {
	// Add one to account for start marker
	return len(l.Data) + 1
}
