package authenticator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsernameClaim_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		kind  claim
		token string
		want  string
	}{
		{
			kind:  NameClaim,
			token: `{"name": "bobby"}`,
			want:  "bobby",
		},
		{
			kind:  EmailClaim,
			token: `{"email": "foo@example.com"}`,
			want:  "foo@example.com",
		},
		{
			kind:  SubClaim,
			token: `{"sub": "111112222"}`,
			want:  "111112222",
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			uc := usernameClaim{kind: tt.kind}
			err := json.Unmarshal([]byte(tt.token), &uc)
			require.NoError(t, err)
			assert.Equal(t, tt.want, uc.value)
		})
	}
}
