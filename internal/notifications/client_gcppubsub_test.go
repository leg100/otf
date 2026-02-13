package notifications

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPubSubClient_New(t *testing.T) {
	// PUBSUB_EMULATOR_HOST is not actually used but ensures the check for valid
	// google credentials is skipped.
	t.Setenv("PUBSUB_EMULATOR_HOST", "foobar")

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
				URL: new(tt.u),
			})
			require.ErrorIs(t, err, tt.want)
		})
	}
}
