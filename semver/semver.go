// Package semver wraps golang.org/x/mod/semver, relaxing the requirement for
// semantic versions to be prefixed with "v".
package semver

import (
	"strings"

	"golang.org/x/mod/semver"
)

func IsValid(s string) bool {
	return semver.IsValid(prefixV(s))
}

func Compare(v, w string) int {
	return semver.Compare(prefixV(v), prefixV(w))
}

func prefixV(s string) string {
	if !strings.HasPrefix(s, "v") {
		s = "v" + s
	}
	return s
}
