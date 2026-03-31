package policy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
)

type cliEvaluator struct {
	logger  logr.Logger
	binPath string
	workDir string
}

func (e *cliEvaluator) Evaluate(ctx context.Context, bundle *MockBundle, policies []*Policy, modules []*PolicyModule) ([]*PolicyCheck, error) {
	root, err := os.MkdirTemp(e.workDir, "otf-sentinel-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)

	for name, data := range bundle.Files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return nil, err
		}
	}

	checks := make([]*PolicyCheck, 0, len(policies))
	for _, policy := range policies {
		config := sentinelConfig()
		for _, mod := range modules {
			if mod.PolicySetID != policy.PolicySetID {
				continue
			}
			moduleName := generatedModuleFilename(mod)
			modulePath := filepath.Join(root, ".otf-modules", moduleName)
			moduleSource := filepath.ToSlash(modulePath)
			config += fmt.Sprintf(`
import "module" %q {
	source = %q
}
`, mod.Name, moduleSource)
			if err := os.MkdirAll(filepath.Dir(modulePath), 0o755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(modulePath, []byte(mod.Source), 0o600); err != nil {
				return nil, err
			}
		}
		if err := os.WriteFile(filepath.Join(root, "sentinel.hcl"), []byte(config), 0o600); err != nil {
			return nil, err
		}
		filename := policy.Name + ".sentinel"
		if policy.Path != "" {
			filename = policy.Path
		}
		policyPath := filepath.Join(root, filepath.FromSlash(filename))
		if err := os.MkdirAll(filepath.Dir(policyPath), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(policyPath, []byte(policy.Source), 0o600); err != nil {
			return nil, err
		}
		cmd := exec.CommandContext(ctx, e.binPath, "apply", "-config=sentinel.hcl", filename)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		output := strings.TrimSpace(string(out))
		passed := err == nil
		if err != nil {
			var exitErr *exec.ExitError
			if !passed && !strings.Contains(output, "false") && !strings.Contains(output, "FAIL") && !strings.Contains(output, "Policy") {
				// Fail closed and surface evaluator/runtime problems directly on the
				// policy check rather than silently dropping checks altogether.
				if !errors.As(err, &exitErr) {
					output = fmt.Sprintf("sentinel execution error: %v", err)
				} else if output == "" {
					output = err.Error()
				}
			}
		}
		checks = append(checks, &PolicyCheck{
			ID:               resource.NewTfeID(resource.PolicyCheckKind),
			PolicySetID:      policy.PolicySetID,
			PolicyID:         policy.ID,
			PolicyName:       policy.Name,
			PolicySetName:    policy.PolicySetName,
			EnforcementLevel: policy.EnforcementLevel,
			Passed:           passed,
			Output:           output,
			CreatedAt:        internal.CurrentTimestamp(nil),
		})
	}
	return checks, nil
}

func generatedModuleFilename(mod *PolicyModule) string {
	name := sentinelIdentifier(mod.Name)
	if name == "" {
		name = "module"
	}
	return name + ".sentinel"
}

func sentinelConfig() string {
	return `
mock "tfrun" {
	module {
		source = "mock-tfrun.sentinel"
	}
}

mock "tfconfig" {
	module {
		source = "mock-tfconfig.sentinel"
	}
}

mock "tfconfig/v1" {
	module {
		source = "mock-tfconfig.sentinel"
	}
}

mock "tfconfig/v2" {
	module {
		source = "mock-tfconfig-v2.sentinel"
	}
}

mock "tfplan" {
	module {
		source = "mock-tfplan.sentinel"
	}
}

mock "tfplan/v1" {
	module {
		source = "mock-tfplan.sentinel"
	}
}

mock "tfplan/v2" {
	module {
		source = "mock-tfplan-v2.sentinel"
	}
}

mock "tfstate" {
	module {
		source = "mock-tfstate.sentinel"
	}
}

mock "tfstate/v1" {
	module {
		source = "mock-tfstate.sentinel"
	}
}

mock "tfstate/v2" {
	module {
		source = "mock-tfstate-v2.sentinel"
	}
}
`
}

func sentinelModuleFromJSON(data []byte) (string, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	return sentinelModuleFromValue(v)
}

func sentinelModuleFromValue(v any) (string, error) {
	switch typed := v.(type) {
	case map[string]any:
		if len(typed) == 0 {
			return "", nil
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var b strings.Builder
		for i, key := range keys {
			if i > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(sentinelIdentifier(key))
			b.WriteString(" = ")
			b.WriteString(sentinelLiteral(typed[key], 0))
			b.WriteByte('\n')
		}
		return b.String(), nil
	default:
		return "value = " + sentinelLiteral(v, 0) + "\n", nil
	}
}

func sentinelIdentifier(s string) string {
	if s == "" {
		return "value"
	}
	var b strings.Builder
	for i, r := range s {
		valid := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (i > 0 && r >= '0' && r <= '9')
		if valid {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "value"
	}
	return out
}

func sentinelLiteral(v any, indent int) string {
	switch typed := v.(type) {
	case nil:
		return "null"
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case string:
		return strconv.Quote(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case []any:
		if len(typed) == 0 {
			return "[]"
		}
		inner := strings.Repeat("  ", indent+1)
		var b strings.Builder
		b.WriteString("[\n")
		for _, item := range typed {
			b.WriteString(inner)
			b.WriteString(sentinelLiteral(item, indent+1))
			b.WriteString(",\n")
		}
		b.WriteString(strings.Repeat("  ", indent))
		b.WriteString("]")
		return b.String()
	case map[string]any:
		if len(typed) == 0 {
			return "{}"
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		inner := strings.Repeat("  ", indent+1)
		var b strings.Builder
		b.WriteString("{\n")
		for _, key := range keys {
			b.WriteString(inner)
			b.WriteString(strconv.Quote(key))
			b.WriteString(": ")
			b.WriteString(sentinelLiteral(typed[key], indent+1))
			b.WriteString(",\n")
		}
		b.WriteString(strings.Repeat("  ", indent))
		b.WriteString("}")
		return b.String()
	default:
		return strconv.Quote(fmt.Sprintf("%v", typed))
	}
}
