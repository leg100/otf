package state

import "sort"

// ChangeAction represents the type of change in a diff.
type ChangeAction string

const (
	ActionAdd    ChangeAction = "add"
	ActionRemove ChangeAction = "remove"
	ActionChange ChangeAction = "change"
)

// ResourceChange is a resource entry in a state diff.
type ResourceChange struct {
	Action   ChangeAction
	Resource Resource
}

// OutputChange is an output entry in a state diff.
type OutputChange struct {
	Action ChangeAction
	Name   string
	Old    *FileOutput
	New    *FileOutput
}

// StateDiff is the diff between two state files.
type StateDiff struct {
	Resources []ResourceChange
	Outputs   []OutputChange
}

// HasChanges returns true when the diff contains at least one change.
func (d StateDiff) HasChanges() bool {
	return len(d.Resources) > 0 || len(d.Outputs) > 0
}

// Diff computes the diff between two state files. from may be nil (first
// version), in which case every entry in to is treated as an addition.
func Diff(from, to *File) StateDiff {
	var d StateDiff

	// --- resources ---
	fromRes := make(map[string]Resource)
	if from != nil {
		for _, r := range from.Resources {
			fromRes[resourceKey(r)] = r
		}
	}
	toRes := make(map[string]Resource)
	for _, r := range to.Resources {
		toRes[resourceKey(r)] = r
	}

	for k, r := range fromRes {
		if _, ok := toRes[k]; !ok {
			d.Resources = append(d.Resources, ResourceChange{Action: ActionRemove, Resource: r})
		}
	}
	for k, r := range toRes {
		if _, ok := fromRes[k]; !ok {
			d.Resources = append(d.Resources, ResourceChange{Action: ActionAdd, Resource: r})
		}
	}
	sort.Slice(d.Resources, func(i, j int) bool {
		return resourceKey(d.Resources[i].Resource) < resourceKey(d.Resources[j].Resource)
	})

	// --- outputs ---
	fromOut := make(map[string]FileOutput)
	if from != nil {
		fromOut = from.Outputs
	}
	toOut := to.Outputs

	allNames := make(map[string]struct{})
	for k := range fromOut {
		allNames[k] = struct{}{}
	}
	for k := range toOut {
		allNames[k] = struct{}{}
	}
	names := make([]string, 0, len(allNames))
	for k := range allNames {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		old, hadOld := fromOut[name]
		newV, hasNew := toOut[name]
		switch {
		case !hadOld && hasNew:
			v := newV
			d.Outputs = append(d.Outputs, OutputChange{Action: ActionAdd, Name: name, New: &v})
		case hadOld && !hasNew:
			v := old
			d.Outputs = append(d.Outputs, OutputChange{Action: ActionRemove, Name: name, Old: &v})
		case hadOld && hasNew:
			if old.StringValue() != newV.StringValue() || old.Sensitive != newV.Sensitive {
				o, n := old, newV
				d.Outputs = append(d.Outputs, OutputChange{Action: ActionChange, Name: name, Old: &o, New: &n})
			}
		}
	}

	return d
}

func resourceKey(r Resource) string {
	return r.Module + "/" + r.Type + "/" + r.Name
}
