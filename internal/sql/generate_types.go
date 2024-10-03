//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

const sqlcConfigPath = "../../sqlc.yaml"

type sqlcConfig struct {
	SQL []struct {
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
	if len(cfg.SQL) != 1 {
		log.Fatal("Error, was expecting only one sqlc engine")
	}
	var types []string
	for _, override := range cfg.SQL[0].Gen.Go.Overrides {
		types = append(types, override.DbType)
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
