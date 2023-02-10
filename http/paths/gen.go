//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

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
)

// action is a controller action
type action struct {
	name       string
	collection bool // whether action acts on collection of resources or a single resource
}

// defaultActions are the default set of actions for a controller of type
// resource
var defaultActions = []action{
	{
		name:       "list",
		collection: true,
	},
	{
		name:       "create",
		collection: true,
	},
	{
		name:       "new",
		collection: true,
	},
	{
		name: "show",
	},
	{
		name: "edit",
	},
	{
		name: "update",
	},
	{
		name: "delete",
	},
}

// controllerSpec is a specification for a controller
type controllerSpec struct {
	Name       string // controller name, used in path names unless path is specified
	prefix     string // prefix path
	nested     []controllerSpec
	path       string
	actions    []action // additional actions
	camel      string
	lowerCamel string

	controllerType
}

type controller struct {
	Name       string
	path       string
	Prefix     string
	Parent     *controller
	Actions    []action // additional paths applying to individual members of collection
	camel      string
	lowerCamel string

	controllerType
}

var specs = []controllerSpec{
	{
		Name:           "login",
		controllerType: singlePath,
	},
	{
		Name:           "logout",
		controllerType: singlePath,
	},
	{
		Name:           "admin_login",
		controllerType: singlePath,
		path:           "/admin/login",
	},
	{
		Name:           "profile",
		controllerType: singlePath,
	},
	{
		Name:           "sessions",
		controllerType: singlePath,
		path:           "/profile/sessions",
	},
	{
		Name:           "revoke_session",
		controllerType: singlePath,
		path:           "/profile/sessions/revoke",
	},
	{
		Name:           "tokens",
		controllerType: singlePath,
		path:           "/profile/tokens",
	},
	{
		Name:           "delete_token",
		controllerType: singlePath,
		path:           "/profile/tokens/delete",
	},
	{
		Name:           "new_token",
		controllerType: singlePath,
		path:           "/profile/tokens/new",
	},
	{
		Name:           "create_token",
		controllerType: singlePath,
		path:           "/profile/tokens/create",
	},
	{
		Name:           "organization",
		controllerType: resourcePath,
		nested: []controllerSpec{
			{
				Name:           "workspace",
				controllerType: resourcePath,
				actions: []action{
					{
						name: "lock",
					},
					{
						name: "unlock",
					},
					{
						name: "set-permission",
					},
					{
						name: "unset-permission",
					},
					{
						name: "watch",
					},
					{
						name: "connect",
					},
					{
						name: "disconnect",
					},
					{
						name: "start-run",
					},
					{
						name: "setup-connection-provider",
					},
					{
						name: "setup-connection-repo",
					},
				},
				nested: []controllerSpec{
					{
						Name:           "run",
						controllerType: resourcePath,
						actions: []action{
							{
								name: "apply",
							},
							{
								name: "discard",
							},
							{
								name: "cancel",
							},
							{
								name: "tail",
							},
						},
					},
					{
						Name:           "variable",
						controllerType: resourcePath,
					},
				},
			},
			{
				Name:           "agent_token",
				controllerType: resourcePath,
			},
			{
				Name:           "user",
				controllerType: resourcePath,
			},
			{
				Name:           "team",
				controllerType: resourcePath,
			},
			{
				Name:           "vcs_provider",
				controllerType: resourcePath,
				camel:          "VCSProvider",
				lowerCamel:     "vcsProvider",
			},
			{
				Name:           "module",
				controllerType: resourcePath,
			},
		},
	},
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
	if r.controllerType == singlePath {
		// single path controllers are just the paths themselves without
		// parameters
		return r.Path()
	}
	var b strings.Builder
	if action.collection {
		if r.Parent != nil {
			b.WriteString(r.Parent.Path())
			b.WriteString("s")
			b.WriteString("/%s")
		}
	}
	b.WriteString(r.Path())
	b.WriteString("s")
	if action.name == "list" {
		// list has no explict action specified in the path
		return b.String()
	}
	b.WriteString("/")
	if action.collection {
		b.WriteString(action.name)
		return b.String()
	}
	b.WriteString("%s")
	if action.name == "show" {
		// show has no explict action specified in the path; show is instead implied using
		// the controller name alone
		return b.String()
	}
	b.WriteString("/")
	b.WriteString(action.name)
	return b.String()
}

// FormatArgs are the args for use with fmt.Sprintf in a path helper in a
// template.
func (r controller) FormatArgs(action action) string {
	return strings.Join(r.params(action.collection), ", ")
}

// HelperName returns the path helper function name for the given action.
func (r controller) HelperName(action action) string {
	switch action.name {
	case "show":
		// show path helper is merely the singular form of the resource name
		return r.Camel()
	case "list":
		// list path helper is merely the plural form of the resource name
		return r.Camel() + "s"
	default:
		return strcase.ToCamel(action.name) + r.Camel()
	}
}

// FuncMapName returns the function map name for the given action.
func (r controller) FuncMapName(action action) string {
	switch action.name {
	case "show":
		// show funcmap name is merely the singular form of the resource name
		return r.LowerCamel() + "Path"
	case "list":
		// list funcmap name is merely the plural form of the resource name
		return r.LowerCamel() + "sPath"
	default:
		// funcmap names for all other actions include their name followed by
		// the resource name
		return strcase.ToLowerCamel(action.name) + r.Camel() + "Path"
	}
}

// HelperParams returns a list of parameters for use in a path helper function
// in a template.
func (r controller) HelperParams(action action) string {
	if params := r.params(action.collection); len(params) > 0 {
		return fmt.Sprintf("%s string", strings.Join(params, ", "))
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

	funcmapTmpl, err := template.New("funcmap.go.tmpl").Funcs(funcmap).ParseFiles("funcmap.go.tmpl")
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

		// Now write to file
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

	// Render single funcmap.go
	if err := funcmapTmpl.Execute(&buf, controllers); err != nil {
		log.Fatal("Error executing template: ", err.Error())
	}

	// Now write to file
	f, err := os.Create("funcmap.go")
	if err != nil {
		log.Fatal("Error:", err.Error())
	}
	_, err = buf.WriteTo(f)
	if err != nil {
		log.Fatal("Error:", err.Error())
	}
	f.Close()
}

// buildControllers recursively builds a slice of controllers
func buildControllers(parent *controller, specs []controllerSpec) []controller {
	var controllers []controller

	for _, spec := range specs {
		ctlr := controller{
			Name:           spec.Name,
			camel:          spec.camel,
			lowerCamel:     spec.lowerCamel,
			path:           spec.path,
			Parent:         parent,
			controllerType: spec.controllerType,
		}
		switch spec.controllerType {
		case resourcePath:
			ctlr.Actions = append(defaultActions, spec.actions...)
		case singlePath:
			ctlr.Actions = []action{{name: "show"}}
		}

		controllers = append(controllers, ctlr)

		if len(spec.nested) > 0 {
			children := buildControllers(&ctlr, spec.nested)
			controllers = append(controllers, children...)
		}
	}
	return controllers
}
