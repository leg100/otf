package otf

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	applyChangesRegex = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.$`)
)

func parseApplyOutput(output string) (Resources, error) {
	matches := applyChangesRegex.FindStringSubmatch(output)
	if matches == nil {
		return Resources{}, fmt.Errorf("regexes unexpectedly did not match apply output")
	}

	adds, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return Resources{}, err
	}
	changes, err := strconv.ParseInt(matches[2], 10, 0)
	if err != nil {
		return Resources{}, err
	}
	deletions, err := strconv.ParseInt(matches[3], 10, 0)
	if err != nil {
		return Resources{}, err
	}

	return Resources{
		ResourceAdditions:    int(adds),
		ResourceChanges:      int(changes),
		ResourceDestructions: int(deletions),
	}, nil
}
