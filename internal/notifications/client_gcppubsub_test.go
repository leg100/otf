package notifications

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestPubSubClient_New(t *testing.T) {
	testutils.SkipIfEnvUnspecified(t, "PUBSUB_EMULATOR_HOST")

	tests := []struct {
		name string
		u    string
		want error
	}{
		{"valid", "gcppubsub://my-project/my-topic", nil},
		{"invalid scheme", "http://my-project/my-topic", ErrInvalidGooglePubSubScheme},
		{"invalid project", "gcppubsub://-my-project/my-topic", ErrInvalidGoogleProjectID},
		{"invalid topic", "gcppubsub://my-project/-my-topic", ErrInvalidGooglePubSubTopic},
		{"missing topic", "gcppubsub://my-project/", ErrInvalidGooglePubSubTopic},
		{"missing trailing /", "gcppubsub://my-project", ErrInvalidGooglePubSubTopic},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newPubSubClient(&Config{
				URL: internal.String(tt.u),
			})
			require.Equal(t, tt.want, err)
		})
	}
}
