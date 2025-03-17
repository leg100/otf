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
	for _, override := range cfg.Overrides.Go.Overrides {
		// Ignore overrides that don't specify a database type.
		if override.DbType == "" {
			continue
		}
		// Ignore overrides of built-in types like int, bool, etc.
		if slices.Contains(ignoreBuiltinTypes, override.DbType) {
			continue
		}
		types = append(types, override.DbType)
		// Add array form, too.
		types = append(types, fmt.Sprintf("%s[]", override.DbType))
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
