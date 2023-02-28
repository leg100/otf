package html

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

const (
	FlashSuccessType flashType = "success"
	FlashWarningType flashType = "warning"
	FlashErrorType   flashType = "error"

	flashCookie = "flash" // name of flash cookie
)

type flashType string

// flash is a flash message for the web UI
type flash struct {
	Type    flashType
	Message string
}

func (f *flash) HTML() template.HTML { return template.HTML(f.Message) }

// PopFlashes pops all flash messages off the stack
func PopFlashes(w http.ResponseWriter, r *http.Request) ([]flash, error) {
	cookie, err := r.Cookie(flashCookie)
	if err != nil {
		// no cookie; return empty stack
		return nil, nil
	}
	decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, err
	}
	var flashes []flash
	if err := json.Unmarshal(decoded, &flashes); err != nil {
		return nil, err
	}
	// purge cookie from browser
	setCookie(w, flashCookie, "", &time.Time{})

	return flashes, nil
}

// flashStack is a stack of flash messages
type flashStack []flash

func (s *flashStack) push(t flashType, msg string) {
	*s = append(*s, flash{t, msg})
}

func (s flashStack) write(w http.ResponseWriter) {
	js, err := json.Marshal(s)
	if err != nil {
		htmlPanic("marshalling flash messages to json: %v", err)
	}
	encoded := base64.URLEncoding.EncodeToString(js)
	setCookie(w, flashCookie, encoded, nil)
}

// FlashSuccess helper writes a single flash success message
func FlashSuccess(w http.ResponseWriter, msg string) {
	flashStack{{Type: FlashSuccessType, Message: msg}}.write(w)
}

// FlashWarning helper writes a single flash warning message
func FlashWarning(w http.ResponseWriter, msg string) {
	flashStack{{Type: FlashWarningType, Message: msg}}.write(w)
}

// FlashError helper writes a single flash error message
func FlashError(w http.ResponseWriter, msg string) {
	flashStack{{Type: FlashErrorType, Message: msg}}.write(w)
}
