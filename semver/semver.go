// Package semver wraps golang.org/x/mod/semver, relaxing the requirement for
// semantic versions to be prefixed with "v".
package semver

import (
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

func IsValid(s string) bool {
	return semver.IsValid(prefixV(s))
}

func Compare(v, w string) int {
	return semver.Compare(prefixV(v), prefixV(w))
}

func Sort(list []string) {
	sort.Sort(semver.ByVersion(list))
}

func prefixV(s string) string {
	if !strings.HasPrefix(s, "v") {
		s = "v" + s
	}
	return s
}

// ByVersion implements sort.Interface for sorting semantic version strings.
type ByVersion []string

func (vs ByVersion) Len() int      { return len(vs) }
func (vs ByVersion) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }
func (vs ByVersion) Less(i, j int) bool {
	cmp := Compare(vs[i], vs[j])
	if cmp != 0 {
		return cmp < 0
	}
	return vs[i] < vs[j]
}
