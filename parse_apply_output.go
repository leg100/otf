package ots

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	applyChangesRegex = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.$`)
)

type apply struct {
	adds, changes, deletions int
}

func parseApplyOutput(output string) (*apply, error) {
	matches := applyChangesRegex.FindStringSubmatch(output)
	if matches == nil {
		return nil, fmt.Errorf("regexes unexpectedly did not match apply output")
	}

	adds, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return nil, err
	}
	changes, err := strconv.ParseInt(matches[2], 10, 0)
	if err != nil {
		return nil, err
	}
	deletions, err := strconv.ParseInt(matches[3], 10, 0)
	if err != nil {
		return nil, err
	}

	return &apply{
		adds:      int(adds),
		changes:   int(changes),
		deletions: int(deletions),
	}, nil
}
