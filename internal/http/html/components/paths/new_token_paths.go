// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"github.com/a-h/templ"
)

func NewToken() templ.SafeURL {
	return templ.URL("/app/current-user/tokens/new")
}
