package http

import (
	"encoding/json"
	"net/http"

	"github.com/leg100/otf"
)

var currentHealthz = Healthz{
	Version: otf.Version,
	Commit:  otf.VCSRef,
	Built:   otf.Built,
}

type Healthz struct {
	Version string
	Commit  string
	Built   string
}

func GetHealthz(w http.ResponseWriter, r *http.Request) {
	payload, err := json.Marshal(currentHealthz)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	w.Header().Set("Content-type", jsonApplication)
	w.Write(payload)
}
