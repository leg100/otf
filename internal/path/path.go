package path

import (
	"strings"

	"github.com/leg100/otf/internal/resource"
)

// Prefix added to all paths
const prefix = "/app"

// Prefix is the site-wide prefix added to all web UI paths requiring authentication.
const Prefix = "/app"

// Resource returns a routable unique path for a single resource with the given
// action and ID.
func Resource(action resource.Action, id resource.ID) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteRune('/')

	b.WriteString(pluralise(id.Kind().Full()))
	b.WriteRune('/')

	b.WriteString(id.String())

	// The 'get' action is special because it is ommitted from the path and it
	// is instead implied.
	if action != resource.Get {
		b.WriteRune('/')
		b.WriteString(action.String())
	}

	return b.String()
}

// Resources returns a routable unique path for a collection of resources of the
// given kind with the given action and parent ID. If the kind doesn't have a parent then
// set parentID to nil.
func Resources(action resource.Action, kind resource.Kind, parentID resource.ID) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteRune('/')

	if parentID != nil {
		b.WriteString(pluralise(parentID.Kind().Full()))
		b.WriteRune('/')

		b.WriteString(parentID.String())
		b.WriteRune('/')
	}

	b.WriteString(pluralise(kind.Full()))

	// The 'list' action is special because it is ommitted from the path and it
	// is instead implied.
	if action != resource.List {
		b.WriteRune('/')
		b.WriteString(action.String())
	}

	return b.String()
}

// Get returns the path for retrieving a single resource.
func Get(id resource.ID) string {
	return Resource(resource.Get, id)
}

// Edit returns the path for editing a resource of the given kind.
func Edit(id resource.ID) string {
	return Resource(resource.Edit, id)
}

// Update returns the path for updating a resource of the given kind.
func Update(id resource.ID) string {
	return Resource(resource.Update, id)
}

// Delete returns the path for deleting a resource of the given kind.
func Delete(id resource.ID) string {
	return Resource(resource.Delete, id)
}

// List returns the path for retrieving a collection of resources.
func List(kind resource.Kind, parentID resource.ID) string {
	return Resources(resource.List, kind, parentID)
}

// New returns the path for constructing a resource of the given kind, with the
// given parent ID.
func New(kind resource.Kind, parentID resource.ID) string {
	return Resources(resource.New, kind, parentID)
}

// Create returns the path for creating a resource of the given kind, with the
// given parent ID.
func Create(kind resource.Kind, parentID resource.ID) string {
	return Resources(resource.Create, kind, parentID)
}

// Add an 's' to make a plural form of the string unless the string
// already ends with an 's'.
func pluralise(s string) string {
	if s[len(s)-1] != 's' {
		return s + "s"
	}
	return s
}
