package otf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type hclOperation func(*hclwrite.File) bool

// RewriteHCL performs HCL surgery on a terraform module.
func RewriteHCL(modulePath string, operations ...hclOperation) error {
	return filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if modulePath != path && info.IsDir() {
			return filepath.SkipDir
		}

		if filepath.Ext(path) != ".tf" {
			return nil
		}

		cfg, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		f, diags := hclwrite.ParseConfig([]byte(cfg), path, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return fmt.Errorf("parsing HCL: %s", diags.Error())
		}

		changed := false
		for _, op := range operations {
			if op(f) {
				changed = true
			}
		}
		if changed {
			if err := os.WriteFile(path, f.Bytes(), 0o644); err != nil {
				return err
			}
		}

		return nil
	})
}

// RemoveBackendBlock is an HCL operation that removes terraform remote backend /
// cloud configuration
func RemoveBackendBlock(f *hclwrite.File) bool {
	for _, block := range f.Body().Blocks() {
		if block.Type() == "terraform" {
			for _, b2 := range block.Body().Blocks() {
				if b2.Type() == "backend" || b2.Type() == "cloud" {
					block.Body().RemoveBlock(b2)
					return true
				}
			}
		}
	}
	return false
}
