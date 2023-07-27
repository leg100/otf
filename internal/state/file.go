package state

import (
	"encoding/json"
	"regexp"
	"strings"
)

var providerPathRegex = regexp.MustCompile(`provider\[".*?/([^"]+)"\]`)

type (
	// File is the terraform state file contents
	File struct {
		Version   int
		Serial    int64
		Lineage   string
		Outputs   map[string]FileOutput
		Resources []Resource
	}

	// FileOutput is an output in the terraform state file
	FileOutput struct {
		Value     json.RawMessage
		Sensitive bool
	}

	Resource struct {
		Name        string
		ProviderURI string `json:"provider"`
		Type        string
		Module      string
	}
)

// Provider extracts the provider from the provider URI
func (r Resource) Provider() string {
	matches := providerPathRegex.FindStringSubmatch(r.ProviderURI)
	if matches == nil || len(matches) < 2 {
		return r.ProviderURI
	}
	return matches[1]
}

func (r Resource) ModuleName() string {
	if r.Module == "" {
		return "root"
	}
	return strings.TrimPrefix(r.Module, "module.")
}

// Type determines the HCL type of the output value
func (r FileOutput) Type() (string, error) {
	var dst any
	if err := json.Unmarshal(r.Value, &dst); err != nil {
		return "", err
	}

	var typ string
	switch dst.(type) {
	case bool:
		typ = "bool"
	case float64:
		typ = "number"
	case string:
		typ = "string"
	case []any:
		typ = "tuple"
	case map[string]any:
		typ = "object"
	case nil:
		typ = "null"
	default:
		typ = "unknown"
	}
	return typ, nil
}

func (r FileOutput) StringValue() string {
	return string(r.Value)
}
