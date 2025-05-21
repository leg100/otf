package forgejo

// related docs: https://forgejo.org/docs/latest/user/webhooks/

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/leg100/otf/internal/vcs"
)

const (
	httpHeaderSignature = "X-Forgejo-Signature"
	httpHeaderEvent     = "X-Forgejo-Event"
)

func HandleEvent(r *http.Request, secret string) (*vcs.EventPayload, error) {
	// verify signature (see forgejo.VerifyWebhookSignatureMiddleware)
	var b bytes.Buffer
	if _, err := io.Copy(&b, r.Body); err != nil {
		return nil, err
	}

	expected := r.Header.Get(httpHeaderSignature)
	if expected == "" {
		return nil, errors.New("no signature found")
	}

	ok, err := forgejo.VerifyWebhookSignature(secret, expected, b.Bytes())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("invalid payload")
	}

	eventtype := r.Header.Get(httpHeaderEvent)

	switch eventtype {
	case "push":
		return handlePushEvent(b.Bytes())
	case "pull_request":
		return handlePullRequestEvent(b.Bytes())
	// forgejo has no "installation" event type.
	default:
		return nil, vcs.NewErrIgnoreEvent("unsupported event: %s", eventtype)
	}
}
