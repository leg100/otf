package helpers

import (
	"bytes"
	"context"
	gohttp "net/http"
	"net/url"

	"github.com/leg100/otf/internal/ui/static"
)

func AssetPath(ctx context.Context, path string) (string, error) {
	return static.AssetsFS.Path(path)
}

func CurrentPath(ctx context.Context) string {
	request := RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.Path
}

func CurrentURL(ctx context.Context) string {
	request := RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.String()
}

func CurrentURLWithoutQuery(ctx context.Context) string {
	request := RequestFromContext(ctx)
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
	FlashSuccess(w, buf.String())
	return nil
}

func Cookie(ctx context.Context, name string) string {
	request := RequestFromContext(ctx)
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
