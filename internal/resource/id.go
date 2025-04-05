package resource

import (
	"fmt"
	"regexp"
)

// ID uniquely identifies an OTF resource.
type ID interface {
	fmt.Stringer
	Kind() Kind
}

// ReStringID is a regular expression used to validate common string ID patterns.
var ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
