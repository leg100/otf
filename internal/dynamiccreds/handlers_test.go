package dynamiccreds

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/testutils"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers(t *testing.T) {
	key, err := jwk.ParseKey(testutils.ReadFile(t, "./testdata/public_key.pem"), jwk.WithPEM(true))
	require.NoError(t, err)

	handlers := &Handlers{publicKey: key}
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handlers.jwks(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode)

	want := `{
  "keys": [
    {
      "e": "AQAB",
      "kid": "B0e3HLS0agadY5a0KDvNhJiSn0e1ZiGFkoH3bqHXH9s",
      "kty": "RSA",
      "n": "snwaHGNZsb97MyX8nec2Km5gSS2nS1kDsH_SrZ2xDInqRcxS1Cbp-L843HCo4eI-XXJPMdHrcliAzgVj9fmcXbWlUenkmOzxRdjPdQtt4rvsTocWb1VeLC2t6Ygvitdn9otYMBTiUgVbPuka-1CW79QhMSIjetRSjbeUeAFh_LSSZdp38dhP7lI2mUP0qbADI94TOUVLOCSfs-LMQqBAJUFn6-Eb75W-HXrofTBuLt19THz_dKLZ4vpsfvOir8FlUeKNMgRCTzZtxR24DfYEdPnUrvK4ToZ30Pp6ZtFb8_9DbRJpuOUIMdEwrijiZIK_lCxiFakrqxHEJaFPMTkh6Hwb7O0ung7yd7L6iYNDKu7RYd0FiTYzt6nUvH6CnJVIHSK9xrUZiYMf9KdKilGig-9Nnw6ttRR7U0IizpvygmujdS3ImJzY2topDbql8xSEYpruBEGPuKJlCDVHjHp_WiH4pVHijcykZQ5WIxDs-sEmukw8RRTtjPtzntOK15vjGin6Bwjtip0f0uWJ4CUQb5Gya29pjQ7z3DASCLdzQnDskkHglvlD-YXsKB8B71VHaoA3xM1iUB75m86T1NX6YGKzFNyW91sdJR36Tmx0ldY8JZnQdqBDCCx4qWGmOjGv2R1W5G-zzAzAswD5e_SPPwzGM7wm5YcXbBosWNemtpk"
    }
  ]
}
`
	assert.Equal(t, want, string(body))
}
