package html

import (
	"reflect"

	"github.com/leg100/otf"
)

// currentOrganization provides an alternative means of determining the current
// organization for the UI helper below.
type currentOrganization interface {
	OrganizationName() string
}

// CurrentOrganization is a UI helper that is used to show the current
// organization on the web page.
func CurrentOrganization(content any) *string {
	// if content implements CurrentOrganization then use that
	if name, ok := content.(currentOrganization); ok {
		return otf.String(name.OrganizationName())
	}
	// otherwise the content has to be a struct with a field named
	// 'Organization' of kind string.
	v := reflect.ValueOf(content)
	// get whatever interface holds
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	// v has to be a struct
	if v.Kind() != reflect.Struct {
		return nil
	}
	// v has to have a field called 'Organization'
	v = v.FieldByName("Organization")
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.String {
		// success
		return otf.String(v.String())
	}
	// no suitable field found.
	return nil
}
