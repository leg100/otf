//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
	"github.com/iancoleman/strcase"
)

type controllerType int

const (
	// resourcePath is a controller with the full complement of restful paths,
	// get, update, new, etc.
	resourcePath controllerType = iota
	// singlePath is a controller with only one path
	singlePath
	// noPath doesn't have any paths; useful for a prefix controller that
	// merely adds a prefix to all nested controllers but doesn't have any paths
	// of its own
	noPath
	// site-wide Prefix added to all routes
	Prefix = "/app"
)

// action is a controller action
type action struct {
	Name       string
	Collection bool `yaml:",omitempty"` // whether action acts on collection of resources or a single resource
}

// defaultActions are the default set of actions for a controller of type
// resource
var defaultActions = []action{
	{
		Name:       "list",
		Collection: true,
	},
	{
		Name:       "create",
		Collection: true,
	},
	{
		Name:       "new",
		Collection: true,
	},
	{
		Name: "show",
	},
	{
		Name: "edit",
	},
	{
		Name: "update",
	},
	{
		Name: "delete",
	},
}

// controllerSpec is a specification for a controller
type controllerSpec struct {
	// controller name, used in path names unless path is specified
	Name   string
	Nested []controllerSpec `yaml:",omitempty"`
	Path   string           `yaml:",omitempty"`
	// additional Actions
	Actions []action `yaml:",omitempty"`
	// whether to skip default set of actions
	SkipDefaultActions bool   `yaml:"skip_default_actions,omitempty"`
	Camel              string `yaml:",camel"`
	LowerCamel         string `yaml:",lower_camel"`
	// disable site-wide prefix
	NoPrefix bool `yaml:"no_prefix,omitempty"`

	ControllerType controllerType `yaml:"controller_type,omitempty"`
}

type controller struct {
	Name   string
	path   string
	Parent *controller
	// additional paths applying to individual members of collection
	Actions []action
	// whether to skip default set of actions
	skipDefaultActions bool
	camel              string
	lowerCamel         string
	// disable site-wide prefix
	noprefix bool

	controllerType
}

func (r controller) Path() string {
	if r.path != "" {
		return r.path
	}
	return "/" + strcase.ToKebab(r.Name)
}

func (r controller) Camel() string {
	if r.camel != "" {
		return r.camel
	}
	return strcase.ToCamel(r.Name)
}

func (r controller) LowerCamel() string {
	if r.lowerCamel != "" {
		return r.lowerCamel
	}
	return strcase.ToLowerCamel(r.Name)
}

// FormatString returns a format string for use with fmt.Sprintf within a
// template for a path helper.
func (r controller) FormatString(action action) string {
	var b strings.Builder
	if !r.noprefix {
		b.WriteString(Prefix)
	}
	if r.controllerType == singlePath {
		// single path controllers are just the paths themselves without
		// parameters
		b.WriteString(r.Path())
		return b.String()
	}
	if action.Collection {
		if r.Parent != nil {
			b.WriteString(r.Parent.Path())
			b.WriteString("s")
			b.WriteString("/%v")
		}
	}
	b.WriteString(r.Path())
	b.WriteString("s")
	if action.Name == "list" {
		// list has no explict action specified in the path
		return b.String()
	}
	b.WriteString("/")
	if action.Collection {
		b.WriteString(action.Name)
		return b.String()
	}
	b.WriteString("%v")
	if action.Name == "show" {
		// show has no explict action specified in the path; show is instead implied using
		// the controller name alone
		return b.String()
	}
	b.WriteString("/")
	b.WriteString(action.Name)
	return b.String()
}

// FormatArgs are the args for use with fmt.Sprintf in a path helper in a
// template.
func (r controller) FormatArgs(action action) string {
	return strings.Join(r.params(action.Collection), ", ")
}

// HelperName returns the path helper function name for the given action.
func (r controller) HelperName(action action) string {
	switch action.Name {
	case "show":
		// show path helper is merely the singular form of the resource name
		return r.Camel()
	case "list":
		// list path helper is merely the plural form of the resource name
		return r.Camel() + "s"
	case "watch":
		if action.Collection {
			return strcase.ToCamel(action.Name) + r.Camel() + "s"
		}
		fallthrough
	default:
		return strcase.ToCamel(action.Name) + r.Camel()
	}
}

// HelperParams returns a list of parameters for use in a path helper function
// in a template.
func (r controller) HelperParams(action action) string {
	if params := r.params(action.Collection); len(params) > 0 {
		return fmt.Sprintf("%s any", strings.Join(params, ", "))
	}
	return ""
}

func (r controller) params(collection bool) []string {
	if r.controllerType == singlePath {
		// single path controllers have no parameters
		return nil
	}
	var params []string
	if collection {
		if r.Parent != nil {
			// only collection actions take the resource id for the parent
			params = append(params, r.Parent.LowerCamel())
		}
	} else {
		// only member actions take a parameter for the controller resource
		params = append(params, r.LowerCamel())
	}
	return params
}

func main() {
	b, err := os.ReadFile("paths.yaml")
	if err != nil {
		log.Fatal("Error reading paths from file: ", err.Error())
	}
	var specs []controllerSpec
	if err := yaml.Unmarshal(b, &specs); err != nil {
		log.Fatal("Error unmarshalling specs: ", err.Error())
	}
	// convert specifications to controllers
	controllers := buildControllers(nil, specs)

	funcmap := template.FuncMap{
		"kebab":      strcase.ToKebab,
		"camel":      strcase.ToCamel,
		"lowerCamel": strcase.ToLowerCamel,
	}
	tmpl, err := template.New("resource.go.tmpl").Funcs(funcmap).ParseFiles("resource.go.tmpl")
	if err != nil {
		log.Fatal("Error parsing template: ", err.Error())
	}

	// Render tmpl out to a tmp buffer first to prevent error messages from
	// being written to files (and to stop files being unnecessarily truncated).
	var buf bytes.Buffer

	// render *_paths.go for each controller
	for _, ctlr := range controllers {
		if err := tmpl.Execute(&buf, ctlr); err != nil {
			log.Fatal("Error executing template: ", err.Error())
		}

		f, err := os.Create(fmt.Sprintf("%s_paths.go", ctlr.Name))
		if err != nil {
			log.Fatal("Error:", err.Error())
		}
		_, err = buf.WriteTo(f)
		if err != nil {
			log.Fatal("Error:", err.Error())
		}
		f.Close()
	}
}

// buildControllers recursively builds a slice of controllers
func buildControllers(parent *controller, specs []controllerSpec) []controller {
	var controllers []controller

	for _, spec := range specs {
		ctlr := controller{
			Name:               spec.Name,
			camel:              spec.Camel,
			lowerCamel:         spec.LowerCamel,
			path:               spec.Path,
			Parent:             parent,
			controllerType:     spec.ControllerType,
			noprefix:           spec.NoPrefix,
			skipDefaultActions: spec.SkipDefaultActions,
		}
		switch spec.ControllerType {
		case resourcePath:
			if ctlr.skipDefaultActions {
				ctlr.Actions = spec.Actions
			} else {
				ctlr.Actions = append(defaultActions, spec.Actions...)
			}
		case singlePath:
			ctlr.Actions = []action{{Name: "show"}}
		}

		controllers = append(controllers, ctlr)

		if len(spec.Nested) > 0 {
			children := buildControllers(&ctlr, spec.Nested)
			controllers = append(controllers, children...)
		}
	}
	return controllers
}
