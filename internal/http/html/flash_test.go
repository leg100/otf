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
	var stack FlashStack
	stack.Push(FlashSuccessType, "yes!")
	stack.Push(FlashWarningType, "uh-oh")
	stack.Push(FlashErrorType, "noooo")
	assert.Equal(t, 3, len(stack))
	w := httptest.NewRecorder()
	stack.Write(w)

	// check content of cookie
	cookies := w.Result().Cookies()
	require.Equal(t, 1, len(cookies))
	want := `[{"Type":"success","Message":"yes!"},{"Type":"warning","Message":"uh-oh"},{"Type":"error","Message":"noooo"}]`
	assert.Equal(t, want, decode64(t, cookies[0].Value))

	// pop flashes
	r := fakeRequest(cookies[0])
	got, err := PopFlashes(r, w)
	require.NoError(t, err)
	require.Equal(t, 3, len(got))
	assert.Contains(t, got, Flash{FlashSuccessType, "yes!"})
	assert.Contains(t, got, Flash{FlashWarningType, "uh-oh"})
	assert.Contains(t, got, Flash{FlashErrorType, "noooo"})
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
