package otf

import "strings"

// ParseBranchRef parses a git ref expecting it to be a reference to a branch. If
// it is not then false is returned, otherwise the branch name along with
// true is returned.
func ParseBranchRef(ref string) (string, bool) {
	parts := strings.Split(ref, "/")
	if len(parts) != 3 {
		return "", false
	}
	if parts[0] == "refs" && parts[1] == "heads" {
		return parts[2], true
	}
	return "", false
}

// ParseRef parses a git ref of the format refs/[tags|heads]/[name],
func ParseRef(ref string) (string, bool) {
	parts := strings.Split(ref, "/")
	if len(parts) != 3 || parts[0] != "refs" {
		return "", false
	}
	if parts[0] == "refs" && parts[1] == "heads" {
		return parts[2], true
	}
	return "", false
}
