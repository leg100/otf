package helpers

import (
	"bytes"
	"context"
	gohttp "net/http"
	"net/url"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
)

func AssetPath(ctx context.Context, path string) (string, error) {
	return html.AssetsFS.Path(path)
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

func IsOwner(ctx context.Context, organization resource.ID) bool {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return false
	}
	if user, ok := subject.(interface {
		IsOwner(resource.ID) bool
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
	request := html.RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.Path
}

func CurrentURL(ctx context.Context) string {
	request := html.RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.String()
}

func CurrentURLWithoutQuery(ctx context.Context) string {
	request := html.RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	// Make a copy of URL to avoid mutating the original.
	u := new(url.URL)
	*u = *request.URL

	u.RawQuery = ""
	return u.String()
}

// TokenFlashMessage is a helper for rendering a flash message with an
// authentication token.
func TokenFlashMessage(w gohttp.ResponseWriter, token []byte) error {
	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := flashToken(string(token)).Render(context.Background(), buf); err != nil {
		return err
	}
	html.FlashSuccess(w, buf.String())
	return nil
}

func Cookie(ctx context.Context, name string) string {
	request := html.RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	for _, cookie := range request.Cookies() {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}
