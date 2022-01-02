package html

import (
	"net/http"
)

type Controller interface {
	// returns struct
	Get(w http.ResponseWriter, r *http.Request)

	// returns []struct
	Index(w http.ResponseWriter, r *http.Request)

	// returns struct
	New(w http.ResponseWriter, r *http.Request)

	// returns status code
	Create(w http.ResponseWriter, r *http.Request)

	// returns struct
	Edit(w http.ResponseWriter, r *http.Request)

	// returns status code
	Update(w http.ResponseWriter, r *http.Request)

	// returns status code
	Delete(w http.ResponseWriter, r *http.Request)
}
