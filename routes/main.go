package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
)

var (
	varRx = regexp.MustCompile(`\{([^\:\}]+)\}`)
)

type app struct {
	otf.Application
}

func main() {
	r := mux.NewRouter()
	html.AddRoutes(logr.Discard(), html.Config{}, &app{}, r)

	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if route.GetName() == "" {
			return nil
		}
		fmt.Printf(route.GetName())
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		matches := varRx.FindAllStringSubmatch(path, -1)
		if matches == nil {
			fmt.Printf("\n")
			return nil
		}
		var vars []string
		for _, m := range matches {
			vars = append(vars, m[1])
		}
		fmt.Printf("(%s)\n", strings.Join(vars, ","))
		return nil
	})
}
