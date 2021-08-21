package agent

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	planChangesRegex = regexp.MustCompile(`(?m)^Plan: (\d+) to add, (\d+) to change, (\d+) to destroy.$`)
	// Either of these messages signals there are no changes (at some point in
	// terraform's history the language changed).
	planNoChangesRegexOld = regexp.MustCompile(`(?m)^No changes. Infrastructure is up-to-date.$`)
	planNoChangesRegex    = regexp.MustCompile(`(?m)^No changes. Your infrastructure matches the configuration.$`)
)

type plan struct {
	adds, changes, deletions int
}

func parsePlanOutput(output string) (*plan, error) {
	if planNoChangesRegex.MatchString(output) || planNoChangesRegexOld.MatchString(output) {
		return &plan{}, nil
	}

	matches := planChangesRegex.FindStringSubmatch(output)
	if matches == nil {
		return nil, fmt.Errorf("regexes unexpectedly did not match plan output")
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

	return &plan{
		adds:      int(adds),
		changes:   int(changes),
		deletions: int(deletions),
	}, nil
}
