package ui

import (
	"encoding/json"
	"reflect"
	"sort"
)

// This file consumes the terraform plan JSON format (format_version 1.x),
// the same file the /api/v2/plans/{plan_id}/json-output endpoint serves, to
// build the view model for the structured plan changes section on the run
// page. Sensitive attribute values are masked at parse time using the
// *_sensitive maps that the format carries alongside the values, so masked
// values never reach the view layer.

// planFile is the subset of the plan JSON format consumed by the view.
type planFile struct {
	ResourceChanges []planResourceChange `json:"resource_changes"`
}

type planResourceChange struct {
	Address string     `json:"address"`
	Change  planChange `json:"change"`
}

type planChange struct {
	Actions []string `json:"actions"`
	Before  map[string]any `json:"before"`
	After   map[string]any `json:"after"`
	// The three fields below are typed `any` because the plan JSON format
	// uses either a bool for whole-object sensitivity/unknown-ness, or a
	// (possibly nested) map for per-attribute sensitivity/unknown-ness.
	AfterUnknown    any `json:"after_unknown"`
	BeforeSensitive any `json:"before_sensitive"`
	AfterSensitive  any `json:"after_sensitive"`
}

// changeView is the view model for one changed resource.
type changeView struct {
	Action  string      // create, update, replace, delete
	Address string      // resource address
	Fields  []fieldView // changed top-level attributes, sorted by key
}

// fieldView is the view model for one changed attribute. Values are already
// rendered (and masked when sensitive).
type fieldView struct {
	Key    string
	Before string // rendered value; empty for created attributes
	After  string // rendered value; "(known after apply)" for unknowns
}

// buildChangeView parses a plan JSON document and returns the view models
// for every resource change, skipping no-op and read-only actions. The
// changes are grouped in TFE order: creates, updates, replaces, destroys.
// The returned slice is non-nil even when empty (no changes), so callers
// can distinguish "unavailable" from "available and empty".
func buildChangeView(planJSON []byte) ([]changeView, error) {
	var pf planFile
	if err := json.Unmarshal(planJSON, &pf); err != nil {
		return nil, err
	}
	views := make([]changeView, 0, len(pf.ResourceChanges))
	for _, rc := range pf.ResourceChanges {
		action := actionString(rc.Change.Actions)
		if action == "" {
			continue // no-op or read: nothing to show
		}
		cv := changeView{Action: action, Address: rc.Address}
		for _, key := range sortedKeys(rc.Change.Before, rc.Change.After) {
			before := rc.Change.Before[key]
			after := rc.Change.After[key]
			// An attribute whose value is only known after apply is a
			// change even though its JSON value is null.
			if reflect.DeepEqual(before, after) && !isUnknown(rc.Change.AfterUnknown, key) {
				continue
			}
			field := fieldView{Key: key}
			if sensitiveAt(rc.Change, key) {
				field.Before = "(sensitive value)"
				field.After = "(sensitive value)"
			} else {
				field.Before = renderValue(before)
				field.After = renderValue(after)
				if isUnknown(rc.Change.AfterUnknown, key) {
					field.After = "(known after apply)"
				}
			}
			cv.Fields = append(cv.Fields, field)
		}
		views = append(views, cv)
	}
	sort.SliceStable(views, func(i, j int) bool {
		ri, rj := actionRank(views[i].Action), actionRank(views[j].Action)
		if ri != rj {
			return ri < rj
		}
		return views[i].Address < views[j].Address
	})
	return views, nil
}

// actionString maps the plan format's actions array onto a display action.
// Read and no-op actions return "" (not shown).
func actionString(actions []string) string {
	switch {
	case len(actions) == 1 && actions[0] == "create":
		return "create"
	case len(actions) == 1 && actions[0] == "update":
		return "update"
	case len(actions) == 1 && actions[0] == "delete":
		return "delete"
	case len(actions) == 2:
		// ["delete", "create"] (replace) or ["create", "delete"]
		return "replace"
	default:
		return ""
	}
}

func actionRank(action string) int {
	switch action {
	case "create":
		return 0
	case "update":
		return 1
	case "replace":
		return 2
	case "delete":
		return 3
	}
	return 4
}

// actionSymbol returns the tally symbol for an action.
func actionSymbol(action string) string {
	switch action {
	case "create":
		return "+"
	case "update":
		return "~"
	case "replace":
		return "+/-"
	case "delete":
		return "-"
	}
	return "?"
}

// actionColor returns the inline color used by the resource report tally.
func actionColor(action string) string {
	switch action {
	case "create":
		return "limegreen"
	case "update":
		return "dodgerblue"
	case "replace":
		return "orange"
	case "delete":
		return "red"
	}
	return "gray"
}

// actionBadge returns the daisyUI badge class for a change row.
func actionBadge(action string) string {
	switch action {
	case "create":
		return "badge-success"
	case "update":
		return "badge-info"
	case "replace":
		return "badge-warning"
	case "delete":
		return "badge-error"
	}
	return "badge-neutral"
}

// countChanges returns the number of changes for a given action.
func countChanges(changes []changeView, action string) int {
	n := 0
	for _, c := range changes {
		if c.Action == action {
			n++
		}
	}
	return n
}

// sortedKeys returns the union of the before/after attribute keys, sorted.
func sortedKeys(before, after map[string]any) []string {
	keys := make(map[string]bool, len(before)+len(after))
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	return sorted
}

// sensitiveAt reports whether the attribute's value is marked sensitive on
// either side of the change. The sensitivity fields carry either a bool
// (whole-object sensitivity: every attribute is sensitive) or a map of
// per-attribute flags (nested maps mask the whole container attribute).
func sensitiveAt(change planChange, key string) bool {
	return truthy(change.BeforeSensitive, key) || truthy(change.AfterSensitive, key)
}

func truthy(v any, key string) bool {
	switch v := v.(type) {
	case bool:
		return v
	case map[string]any:
		switch s := v[key].(type) {
		case bool:
			return s
		case map[string]any:
			// container with sensitive children: mask the whole attribute
			return len(s) > 0
		}
	}
	return false
}

// isUnknown reports whether the after value is known only after apply. The
// unknown field carries either a bool (whole object) or a per-attribute map.
func isUnknown(afterUnknown any, key string) bool {
	switch v := afterUnknown.(type) {
	case bool:
		return v
	case map[string]any:
		if b, ok := v[key].(bool); ok {
			return b
		}
	}
	return false
}

// renderValue renders an attribute value for display: strings verbatim,
// everything else compact JSON, and nil as an empty string (so created
// attributes show no before value, and destroyed attributes no after one).
func renderValue(v any) string {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "(unrenderable)"
		}
		return string(b)
	}
}
