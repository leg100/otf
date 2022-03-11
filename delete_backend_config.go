package otf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func deleteBackendConfigFromDirectory(ctx context.Context, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if dir != path && info.IsDir() {
			return filepath.SkipDir
		}

		if filepath.Ext(path) != ".tf" {
			return nil
		}

		in, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		deleted, out, err := deleteBackendConfig(in)
		if err != nil {
			return nil
		}

		if deleted {
			if err := os.WriteFile(path, out, 0644); err != nil {
				return err
			}
		}

		return nil
	})
}

func deleteBackendConfig(config []byte) (bool, []byte, error) {
	f, diags := hclwrite.ParseConfig([]byte(config), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return false, nil, fmt.Errorf("unable to parse HCL: %s", diags.Error())
	}

	for _, block := range f.Body().Blocks() {
		if block.Type() == "terraform" {
			for _, b2 := range block.Body().Blocks() {
				if b2.Type() == "backend" {
					block.Body().RemoveBlock(b2)
					return true, f.Bytes(), nil
				}
			}
		}
	}

	return false, f.Bytes(), nil
}
