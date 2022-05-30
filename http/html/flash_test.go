package html

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlash(t *testing.T) {
	r := &http.Request{}
	store := &flashStore{
		db:       make(map[string]flash),
		getToken: func(context.Context) string { return "token123" },
	}
	want := flashSuccess("great news")
	store.push(r, want)
	got := store.popFunc(r)()
	if assert.NotNil(t, got) {
		assert.Equal(t, want, *got)
	}
	got = store.popFunc(r)()
	assert.Nil(t, got)
}
