package html

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlash(t *testing.T) {
	// write flashes
	var stack flashStack
	stack.push(FlashSuccessType, "yes!")
	stack.push(FlashWarningType, "uh-oh")
	stack.push(FlashErrorType, "noooo")
	assert.Equal(t, 3, len(stack))
	w := httptest.NewRecorder()
	stack.write(w)

	// check content of cookie
	cookies := w.Result().Cookies()
	require.Equal(t, 1, len(cookies))
	want := `[{"Type":"success","Message":"yes!"},{"Type":"warning","Message":"uh-oh"},{"Type":"error","Message":"noooo"}]`
	assert.Equal(t, want, decode64(t, cookies[0].Value))

	// pop flashes
	r := fakeRequest(cookies[0])
	w = httptest.NewRecorder()
	got, err := PopFlashes(w, r)
	require.NoError(t, err)
	require.Equal(t, 3, len(got))
	assert.Contains(t, got, flash{FlashSuccessType, "yes!"})
	assert.Contains(t, got, flash{FlashWarningType, "uh-oh"})
	assert.Contains(t, got, flash{FlashErrorType, "noooo"})

	// cookie should now be set to be purged
	cookies = w.Result().Cookies()
	require.Equal(t, 1, len(cookies))
	assert.Equal(t, -1, cookies[0].MaxAge)
}

func TestFlashHelpers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		FlashSuccess(w, "yes!")

		cookies := w.Result().Cookies()
		require.Equal(t, 1, len(cookies))

		want := `[{"Type":"success","Message":"yes!"}]`
		assert.Equal(t, want, decode64(t, cookies[0].Value))
	})

	t.Run("warning", func(t *testing.T) {
		w := httptest.NewRecorder()
		FlashWarning(w, "uh-oh")

		cookies := w.Result().Cookies()
		require.Equal(t, 1, len(cookies))

		want := `[{"Type":"warning","Message":"uh-oh"}]`
		assert.Equal(t, want, decode64(t, cookies[0].Value))
	})

	t.Run("error", func(t *testing.T) {
		w := httptest.NewRecorder()
		FlashError(w, "noooo")

		cookies := w.Result().Cookies()
		require.Equal(t, 1, len(cookies))

		want := `[{"Type":"error","Message":"noooo"}]`
		assert.Equal(t, want, decode64(t, cookies[0].Value))
	})
}

func decode64(t *testing.T, encoded string) string {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	require.NoError(t, err)
	return string(decoded)
}

func fakeRequest(cookie *http.Cookie) *http.Request {
	r := &http.Request{}
	r.Header = make(map[string][]string)
	r.AddCookie(cookie)
	return r
}
