package html

import (
	"html/template"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
)

var FuncMap = template.FuncMap{}

func init() {
	// template functions
	FuncMap = sprig.HtmlFuncMap()
	// make version available to templates
	FuncMap["version"] = func() string { return otf.Version }
	// make version available to templates
	FuncMap["trimHTML"] = func(tmpl template.HTML) template.HTML { return template.HTML(strings.TrimSpace(string(tmpl))) }
	FuncMap["mergeQuery"] = mergeQuery
	FuncMap["selected"] = selected
	FuncMap["checked"] = checked
	FuncMap["disabled"] = disabled
	// make path helpers available to templates
	for k, v := range paths.FuncMap() {
		FuncMap[k] = v
	}
}
