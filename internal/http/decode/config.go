package decode

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

var ErrDecodeConfig = errors.New("config could not be decoded as either a query string nor as JSON")

// DecodeConfig decodes the config contained in src to the struct in dst. Src is
// expected to be either JSON encoded or to be a URL query string.
func DecodeConfig(dst any, src []byte) error {
	// First assume src is a query string
	q, err := url.ParseQuery(string(src))
	if err != nil {
		return err
	}
	queryStringErr := Decode(dst, q)
	if queryStringErr == nil {
		// Successfully decoded query string
		return nil
	}
	// Not a query string; try unmarshalling as json instead
	jsonErr := json.Unmarshal(src, dst)
	if jsonErr == nil {
		// Successfully decoded query string
		return nil
	}
	return fmt.Errorf("%w: %w: %w", ErrDecodeConfig, queryStringErr, jsonErr)
}
