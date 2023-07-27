package internal

import (
	"reflect"
	"regexp"
	"strings"
)

// ReStringID is a regular expression used to validate common string ID patterns.
var ReStringID = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)

// GetID retrieves the ID field of a struct contained in s. If s is not a struct,
// or there is no ID field, then false is returned.
func GetID(s any) (string, bool) {
	v := reflect.Indirect(reflect.ValueOf(s))
	if v.Kind() != reflect.Struct {
		return "", false
	}
	f := v.FieldByName("ID")
	if !f.IsValid() {
		return "", false
	}
	return f.String(), true
}

// NewID constructs resource IDs, which are composed of the resource type and a
// random 16 character string, separated by a hyphen.
func NewID(rtype string) string {
	return rtype + "-" + GenerateRandomString(16)
}

// ValidStringID checks if the given string pointer is non-nil and
// contains a typical string identifier.
func ValidStringID(v *string) bool {
	return v != nil && ReStringID.MatchString(*v)
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
