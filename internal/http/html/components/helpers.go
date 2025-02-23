package components

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/html"
)

func AssetPath(ctx context.Context, path string) (string, error) {
	if fs := html.AssetsFS(ctx); fs != nil {
		return fs.Path(path)
	}
	return "", errors.New("not found")
}

func CurrentUsername(ctx context.Context) (string, error) {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return "", err
	}
	return subject.String(), nil
}

func Authenticated(ctx context.Context) bool {
	if _, err := authz.SubjectFromContext(ctx); err != nil {
		return false
	}
	return true
}

func IsOwner(ctx context.Context, organization string) bool {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return false
	}
	if user, ok := subject.(interface {
		IsOwner(string) bool
	}); ok {
		return user.IsOwner(organization)
	}
	return false
}

func IsSiteAdmin(ctx context.Context) bool {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return false
	}
	if user, ok := subject.(interface {
		IsSiteAdmin() bool
	}); ok {
		return user.IsSiteAdmin()
	}
	return false
}

func CurrentPath(ctx context.Context) string {
	request := http.RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.Path
}

func CurrentURL(ctx context.Context) string {
	request := http.RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.String()
}
