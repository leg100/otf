package run

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/leg100/otf"
)

var applyChangesRegex = regexp.MustCompile(`(?m)^Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed.`)

func ParseApplyOutput(output string) (otf.ResourceReport, error) {
	matches := applyChangesRegex.FindStringSubmatch(output)
	if matches == nil {
		return otf.ResourceReport{}, fmt.Errorf("regexes unexpectedly did not match apply output")
	}

	adds, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return otf.ResourceReport{}, err
	}
	changes, err := strconv.ParseInt(matches[2], 10, 0)
	if err != nil {
		return otf.ResourceReport{}, err
	}
	deletions, err := strconv.ParseInt(matches[3], 10, 0)
	if err != nil {
		return otf.ResourceReport{}, err
	}

	return otf.ResourceReport{
		Additions:    int(adds),
		Changes:      int(changes),
		Destructions: int(deletions),
	}, nil
}
