//go:build ignore

//
// This go generation script generates a list of postgres table types. It parses
// the sqlc.yaml config file, inferring from the list of overridden types the
// list of postgres table types, and then writes them out to a go string slice,
// adding both the singular and array forms of the table.

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"slices"
	"text/template"

	"gopkg.in/yaml.v3"
)

const sqlcConfigPath = "../../sqlc.yaml"

var ignoreBuiltinTypes = []string{
	"pg_catalog.bool",
	"bool",
	"pg_catalog.int4",
	"pg_catalog.int8",
	"text",
}

type sqlcConfig struct {
	Overrides struct {
		Go struct {
			Overrides []struct {
				DbType string `yaml:"db_type"`
			}
		}
	} `yaml:"overrides"`
	Sql []struct {
		Gen struct {
			Go struct {
				Overrides []struct {
					DbType string `yaml:"db_type"`
				}
			}
		}
	} `yaml:"sql"`
}

func main() {
	sqlcConfigFile, err := os.ReadFile(sqlcConfigPath)
	if err != nil {
		log.Fatal("Error reading sqlc config file: ", err.Error())
	}
	var cfg sqlcConfig
	if err := yaml.Unmarshal(sqlcConfigFile, &cfg); err != nil {
		log.Fatal("Error unmarshaling sqlc config file: ", err.Error())
	}
	var types []string
	addType := func(dbType string) {
		// Ignore overrides that don't specify a database type.
		if dbType == "" {
			return
		}
		// Ignore overrides of built-in types like int, bool, etc.
		if slices.Contains(ignoreBuiltinTypes, dbType) {
			return
		}
		types = append(types, dbType)
		// Add array form, too.
		types = append(types, fmt.Sprintf("%s[]", dbType))
	}
	// Add global override types
	for _, override := range cfg.Overrides.Go.Overrides {
		addType(override.DbType)
	}
	// Add per-table override types
	for _, table := range cfg.Sql {
		for _, override := range table.Gen.Go.Overrides {
			addType(override.DbType)
		}
	}
	tmpl, err := template.New("types.go.tmpl").ParseFiles("types.go.tmpl")
	if err != nil {
		log.Fatal("Error parsing template: ", err.Error())
	}
	// Render tmpl out to a tmp buffer first to prevent error messages from
	// being written to files (and to stop files being unnecessarily truncated).
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, types); err != nil {
		log.Fatal("Error executing template: ", err.Error())
	}
	// Now write to file
	f, err := os.Create("types.go")
	if err != nil {
		log.Fatal("Error:", err.Error())
	}
	defer f.Close()
	if _, err = buf.WriteTo(f); err != nil {
		log.Fatal("Error:", err.Error())
	}
}
