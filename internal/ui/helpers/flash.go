package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
)

const (
	FlashSuccessType FlashType = "success"
	FlashWarningType FlashType = "warning"
	FlashErrorType   FlashType = "error"

	flashCookie = "flash" // name of flash cookie
)

var flashColors = map[FlashType]string{
	FlashSuccessType: "alert-success",
	FlashWarningType: "alert-warning",
	FlashErrorType:   "alert-error",
}

var flashIcons = map[FlashType]templ.Component{
	FlashSuccessType: successIcon(),
	FlashWarningType: warningIcon(),
	FlashErrorType:   errorIcon(),
}

type FlashType string

func (f FlashType) String() string { return string(f) }

// Flash is a Flash message for the web UI
type Flash struct {
	Type    FlashType
	Message string
}

// PopFlashes pops all flash messages off the stack
func PopFlashes(r *http.Request, w http.ResponseWriter) ([]Flash, error) {
	cookie, err := r.Cookie(flashCookie)
	if err != nil {
		// no cookie; return empty stack
		return nil, nil
	}
	decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, err
	}
	var flashes []Flash
	if err := json.Unmarshal(decoded, &flashes); err != nil {
		return nil, err
	}
	// Purge cookie from browser.
	SetCookie(w, flashCookie, "", &time.Time{})
	return flashes, nil
}

// FlashStack is a stack of flash messages
type FlashStack []Flash

func (s *FlashStack) Push(t FlashType, msg string) {
	*s = append(*s, Flash{t, msg})
}

func (s FlashStack) Write(w http.ResponseWriter) {
	js, err := json.Marshal(s)
	if err != nil {
		// upstream middleware catches the panic
		panic(fmt.Sprintf("marshalling flash messages to json: %v", err))
	}
	encoded := base64.URLEncoding.EncodeToString(js)
	SetCookie(w, flashCookie, encoded, nil)
}

// FlashSuccess helper writes a single flash success message
func FlashSuccess(w http.ResponseWriter, msg string) {
	FlashStack{{Type: FlashSuccessType, Message: msg}}.Write(w)
}

// FlashWarning helper writes a single flash warning message
func FlashWarning(w http.ResponseWriter, msg string) {
	FlashStack{{Type: FlashWarningType, Message: msg}}.Write(w)
}

// FlashError helper writes a single flash error message
func FlashError(w http.ResponseWriter, msg string) {
	FlashStack{{Type: FlashErrorType, Message: msg}}.Write(w)
}
