package resource

import (
	"regexp"
	"strings"

	"github.com/leg100/otf/internal"
)

var (
	// ReStringID is a regular expression used to validate common string ID patterns.
	ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
	// base58 alphabet
	base58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

type (
	ID struct {
		Kind Kind
		ID   string
	}

	Kind string
)

// NewID constructs resource IDs, composed of:
// (1) a symbol representing a resource type, e.g. "ws" for workspaces
// (2) a hyphen
// (3) a 16 character string composed of random characters from the base58 alphabet
func NewID(rtype string) string {
	return rtype + "-" + internal.GenerateRandomStringFromAlphabet(16, base58)
}

// ConvertID converts an ID for use with a different resource, e.g. convert
// run-123 to plan-123.
func ConvertID(id, resource string) string {
	parts := strings.Split(id, "-")
	// if ID not in expected form then just return it unchanged without error
	if len(parts) != 2 {
		return id
	}
	return resource + "-" + parts[1]
}
