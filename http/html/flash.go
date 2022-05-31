package html

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

const (
	FlashSuccess flashType = "success"
	FlashError   flashType = "error"
	// name of flash cookie
	flashCookie = "flash"
)

// flash represents a flash message indicating success/failure to a user
type flash struct {
	Type    flashType
	Message string
}

type flashType string

// setFlash sets flash message on response cookie
func setFlash(w http.ResponseWriter, f flash) {
	js, err := json.Marshal(f)
	if err != nil {
		// reliant on middleware catching panic and sending HTTP500 to user
		panic("marshalling flash message to json: " + err.Error())
	}
	encoded := base64.URLEncoding.EncodeToString(js)
	setCookie(w, flashCookie, encoded, nil)
}

// flashSuccess helper
func flashSuccess(w http.ResponseWriter, msg string) {
	setFlash(w, flash{Type: FlashSuccess, Message: msg})
}

// flashError helper
func flashError(w http.ResponseWriter, msg string) {
	setFlash(w, flash{Type: FlashError, Message: msg})
}

// popFlashFunc returns a func to pop a flash message for the current session - for
// use in a go template
func popFlashFunc(w http.ResponseWriter, r *http.Request) func() *flash {
	c, err := r.Cookie(flashCookie)
	if err != nil {
		// err should only ever be http.ErrNoCookie
		return func() *flash { return nil }
	}
	value, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		// reliant on middleware catching panic and sending HTTP500 to user
		panic("decoding flash message: " + err.Error())
	}
	var f flash
	if err := json.Unmarshal(value, &f); err != nil {
		// reliant on middleware catching panic and sending HTTP500 to user
		panic("unmarshalling flash message: " + err.Error())
	}
	setCookie(w, flashCookie, "", &time.Time{})
	return func() *flash { return &f }
}
