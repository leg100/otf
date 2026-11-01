package ui

import (
	"regexp"
	"strings"
)

// This file parses the human-readable apply log into the view model for the
// structured apply changes section. Applies have no JSON equivalent of the
// plan file, but the engine's log is stable enough to extract per-resource
// operations: "<address>: Creating...", "<address>: Creation complete after
// 2s [id=...]". This is the same approach the domain layer uses for the
// apply tally (run.ParseApplyOutput), extended to per-resource entries.

// applyChangeView is the view model for one applied resource operation.
type applyChangeView struct {
	Address  string // resource address
	Action   string // create, update, destroy
	Duration string // e.g. "2s"; empty until complete
	ID       string // resource id, optional
	Complete bool   // whether the operation completed
}

// ANSI escape sequences are stripped before matching: the raw log chunks
// carry terminal colors, which the log pane renders via terminal-to-html.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

var applyStartRe = regexp.MustCompile(`^(.+?): (Creating|Modifying|Destroying)\.\.\.(?: \[id=([^\]]+)\])?$`)
var applyCompleteRe = regexp.MustCompile(`^(.+?): (Creation|Modifications|Destruction) complete after (\S+)(?: \[id=([^\]]+)\])?$`)

// parseApplyView parses an apply log and returns one view model per
// resource operation, in completion order. In-progress operations (started
// but not yet complete) are included with Complete=false. Progress lines
// ("Still creating...") and non-operation lines are skipped.
func parseApplyView(log string) []applyChangeView {
	changes := []applyChangeView{}
	for _, line := range strings.Split(ansiRe.ReplaceAllString(log, ""), "\n") {
		line = strings.TrimRight(line, "\r")
		if m := applyStartRe.FindStringSubmatch(line); m != nil {
			changes = append(changes, applyChangeView{
				Address: m[1],
				Action:  applyStartAction(m[2]),
				ID:      m[3],
			})
			continue
		}
		if m := applyCompleteRe.FindStringSubmatch(line); m != nil {
			action := applyCompleteAction(m[2])
			// attach to the most recent matching start that is not yet
			// complete (a replace emits a destroy and a create for the
			// same address, in order)
			for i := len(changes) - 1; i >= 0; i-- {
				if changes[i].Address == m[1] && changes[i].Action == action && !changes[i].Complete {
					changes[i].Duration = m[3]
					changes[i].Complete = true
					if m[4] != "" {
						changes[i].ID = m[4]
					}
					break
				}
			}
		}
	}
	return changes
}

func applyStartAction(op string) string {
	switch op {
	case "Creating":
		return "create"
	case "Modifying":
		return "update"
	case "Destroying":
		return "destroy"
	}
	return ""
}

func applyCompleteAction(op string) string {
	switch op {
	case "Creation":
		return "create"
	case "Modifications":
		return "update"
	case "Destruction":
		return "destroy"
	}
	return ""
}

// countApplyChanges returns the number of operations for a given action.
func countApplyChanges(changes []applyChangeView, action string) int {
	n := 0
	for _, c := range changes {
		if c.Action == action {
			n++
		}
	}
	return n
}
