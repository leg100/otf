package otf

import "strings"

// ParseBranch parses a git ref expecting it to be a reference to a branch. If
// it is not then false is returned, otherwise the branch name along with
// true is returned.
func ParseBranch(ref string) (string, bool) {
	parts := strings.Split(ref, "/")
	if len(parts) != 3 {
		return "", false
	}
	if parts[0] == "refs" && parts[1] == "heads" {
		return parts[2], true
	}
	return "", false
}
