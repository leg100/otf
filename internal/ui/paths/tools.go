package paths

// import packages here that the generator (gen.go) needs because the generator
// has a build constraint which means it is ignored when running stuff like "go
// mod tidy".
import (
	_ "github.com/goccy/go-yaml"
	_ "github.com/iancoleman/strcase"
)
