// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"github.com/a-h/templ"
)

func DeleteToken() templ.SafeURL {
	return templ.URL("/app/current-user/tokens/delete")
}
