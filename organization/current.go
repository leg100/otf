package organization

import (
	"reflect"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
)

func init() {
	html.FuncMap["currentOrganization"] = CurrentOrganization
}

// CurrentOrganization is a UI helper that is used to show the current
// organization in the header of a web page.
func CurrentOrganization(content any) *string {
	// if content *is* an organization then use that
	if org, ok := content.(*Organization); ok {
		return &org.Name
	}
	v := reflect.ValueOf(content)
	// get whatever a pointer points to, or interface holds
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
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
	// dereference v if a pointer
	v = reflect.Indirect(v)
	switch v.Kind() {
	case reflect.String:
		// struct has a field of kind string called 'Organization', which we
		// infer to be the current organization.
		return otf.String(v.String())
	case reflect.Struct:
		// struct has a field of kind struct, which must be an Organization
		// to be inferred as the current organization
		if org, ok := v.Interface().(Organization); ok {
			return &org.Name
		}
		return nil
	default:
		return nil
	}
}
