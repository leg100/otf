// Package tfe provides common functionality useful for implementation of the
// Hashicorp TFE/TFC API, which uses the json:api encoding
package tfe

import (
	"io"

	"github.com/DataDog/jsonapi"
)

func Unmarshal(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return jsonapi.Unmarshal(b, v)
}
