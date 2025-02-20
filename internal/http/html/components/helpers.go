package components

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal/http/html"
)

func assetPath(ctx context.Context, path string) (string, error) {
	if fs := html.AssetsFS(ctx); fs != nil {
		return fs.Path(path)
	}
	return "", errors.New("not found")
}
