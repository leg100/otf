package otf

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	applyChangesRegex = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.$`)
)

func ParseApplyOutput(output string) (ResourceReport, error) {
	matches := applyChangesRegex.FindStringSubmatch(output)
	if matches == nil {
		return ResourceReport{}, fmt.Errorf("regexes unexpectedly did not match apply output")
	}

	adds, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return ResourceReport{}, err
	}
	changes, err := strconv.ParseInt(matches[2], 10, 0)
	if err != nil {
		return ResourceReport{}, err
	}
	deletions, err := strconv.ParseInt(matches[3], 10, 0)
	if err != nil {
		return ResourceReport{}, err
	}

	return ResourceReport{
		Additions:    int(adds),
		Changes:      int(changes),
		Destructions: int(deletions),
	}, nil
}
