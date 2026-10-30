package run

import (
	"fmt"
	"regexp"
	"strconv"
)

// The imported count is only printed when the apply imported anything, e.g.
// "Apply complete! Resources: 1 imported, 1 added, 0 changed, 0 destroyed."
var applyChangesRegex = regexp.MustCompile(`(?m)^Apply complete! Resources: (?:\d+ imported, )?(\d+) added, (\d+) changed, (\d+) destroyed.`)

func ParseApplyOutput(output string) (Report, error) {
	matches := applyChangesRegex.FindStringSubmatch(output)
	if matches == nil {
		return Report{}, fmt.Errorf("regexes unexpectedly did not match apply output")
	}

	adds, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return Report{}, err
	}
	changes, err := strconv.ParseInt(matches[2], 10, 0)
	if err != nil {
		return Report{}, err
	}
	deletions, err := strconv.ParseInt(matches[3], 10, 0)
	if err != nil {
		return Report{}, err
	}

	return Report{
		Additions:    int(adds),
		Changes:      int(changes),
		Destructions: int(deletions),
	}, nil
}
