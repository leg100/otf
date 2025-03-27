package resource

import (
	"fmt"
	"regexp"
)

type ID interface {
	fmt.Stringer
}

// ReStringID is a regular expression used to validate common string ID patterns.
var ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
