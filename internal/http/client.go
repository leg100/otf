package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/DataDog/jsonapi"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
)

type (
	Client struct {
		Token string

		baseURL *url.URL
		headers http.Header
		http    *retryablehttp.Client
	}

	// ClientConfig provides configuration details to the API client.
	ClientConfig struct {
		// The URL of the otf API.
		URL string
		// The base path on which the API is served.
		BasePath string
		// API token used to access the otf API.
		Token string
		// Headers that will be added to every request.
		Headers http.Header
		// Toggle retrying requests upon encountering transient errors.
		RetryRequests bool
		// Override default http transport
		Transport http.RoundTripper
		// Logger for logging an error upon retry
		Logger logr.Logger
	}
)

func NewClient(config ClientConfig) (*Client, error) {
	// set defaults
	if config.URL == "" {
		config.URL = DefaultURL
	}
	if config.BasePath == "" {
		config.BasePath = APIBasePath
	}
	if config.Headers == nil {
		config.Headers = make(http.Header)
	}
	if config.Transport == nil {
		config.Transport = http.DefaultTransport
	}
	config.Headers.Set("User-Agent", "otf-agent")

	baseURL, err := ParseURL(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid server url: %v", err)
	}
	baseURL.Path = path.Join(baseURL.Path, config.BasePath)
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	// This value must be provided by the user.
	if config.Token == "" {
		return nil, fmt.Errorf("missing API token")
	}

	// Create the client.
	client := &Client{
		baseURL: baseURL,
		Token:   config.Token,
		headers: config.Headers,
	}
	client.http = &retryablehttp.Client{
		Backoff:      retryablehttp.DefaultBackoff,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   &http.Client{Transport: config.Transport},
		RetryWaitMin: 500 * time.Millisecond,
		RetryWaitMax: 30 * time.Second,
		RetryMax:     30,
	}
	if config.RetryRequests {
		// enable retries
		client.http.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			retry, retryErr := retryablehttp.ErrorPropagatedRetryPolicy(ctx, resp, err)
			if retry {
				// Log retry attempts. The ErrorPropagatedRetryPolicy sometimes
				// returns an error explaining why it has decided to retry and
				// if so then report this error rather than the original error.
				if retryErr != nil {
					err = retryErr
				}
				// The http response is nil when there is a problem with the
				// request and there is no response, e.g. socket timeout.
				if resp != nil && resp.Request != nil {
					// Try to decode more informative error message from
					// response body.
					if jsonapiErr := tryUnmarshalJSONAPIError(resp.Body); jsonapiErr != nil {
						err = jsonapiErr
					}
					config.Logger.Error(err, "retrying request", "url", resp.Request.URL, "status", resp.StatusCode)
				} else {
					config.Logger.Error(err, "retrying request")
				}
			}
			return retry, retryErr
		}
	} else {
		// disable retries
		client.http.CheckRetry = func(_ context.Context, _ *http.Response, err error) (bool, error) {
			return false, err
		}
	}
	return client, nil
}

// Hostname returns the server host:port.
func (c *Client) Hostname() string {
	return c.baseURL.Host
}

// NewRequest creates an API request with proper headers and serialization.
//
// A relative URL path can be provided, in which case it is resolved relative to the baseURL
// of the Client. Relative URL paths should always be specified without a preceding slash. Adding a
// preceding slash allows for ignoring the configured baseURL for non-standard endpoints.
//
// If v is supplied, the value will be JSONAPI encoded and included as the
// request body. If the method is GET, the value will be parsed and added as
// query parameters.
func (c *Client) NewRequest(method, path string, v any) (*retryablehttp.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Authorization", "Bearer "+c.Token)

	var body any
	switch method {
	case "GET":
		reqHeaders.Set("Accept", "application/vnd.api+json")

		if v != nil {
			q := url.Values{}
			if err := Encoder.Encode(v, q); err != nil {
				return nil, err
			}
			u.RawQuery = q.Encode()
		}
	case "DELETE", "PATCH", "POST":
		reqHeaders.Set("Accept", "application/vnd.api+json")
		reqHeaders.Set("Content-Type", "application/vnd.api+json")

		if v != nil {
			if body, err = serializeRequestBody(v); err != nil {
				return nil, err
			}
		}
	case "PUT":
		reqHeaders.Set("Accept", "application/json")
		reqHeaders.Set("Content-Type", "application/octet-stream")
		body = v
	}

	req, err := retryablehttp.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Set the default headers.
	maps.Copy(req.Header, c.headers)

	// Set the request specific headers.
	maps.Copy(req.Header, reqHeaders)

	return req, nil
}

// Helper method that serializes the given ptr or ptr slice into a JSON
// request. It automatically uses jsonapi or json serialization, depending
// on the body type's tags.
func serializeRequestBody(v any) (any, error) {
	// The body can be a slice of pointers or a pointer. In either
	// case we want to choose the serialization type based on the
	// individual record type. To determine that type, we need
	// to either follow the pointer or examine the slice element type.
	// There are other theoretical possiblities (e. g. maps,
	// non-pointers) but they wouldn't work anyway because the
	// json-api library doesn't support serializing other things.
	var modelType reflect.Type
	bodyType := reflect.TypeOf(v)
	invalidBodyError := errors.New("DELETE/PATCH/POST body must be nil, ptr, or ptr slice")
	switch bodyType.Kind() {
	case reflect.Slice:
		sliceElem := bodyType.Elem()
		if sliceElem.Kind() != reflect.Pointer {
			return nil, invalidBodyError
		}
		modelType = sliceElem.Elem()
	case reflect.Pointer:
		modelType = reflect.ValueOf(v).Elem().Type()
	default:
		return nil, invalidBodyError
	}

	// Infer whether the request uses jsonapi or regular json
	// serialization based on how the fields are tagged.
	jsonAPIFields := 0
	for structField := range modelType.Fields() {
		if structField.Tag.Get("jsonapi") != "" {
			jsonAPIFields++
		}
	}

	// If there is at least one field tagged with jsonapi then use the jsonapi
	// marshaler.
	if jsonAPIFields > 0 {
		return jsonapi.Marshal(v, jsonapi.MarshalClientMode())
	} else {
		return json.Marshal(v)
	}
}

// Do sends an API request and returns the API response. The API response
// is JSONAPI decoded and the document's primary data is stored in the value
// pointed to by v, or returned as an error if an API error has occurred.
//
// If v implements the io.Writer interface, the raw response body will be
// written to v, without attempting to first decode it.
//
// The provided ctx must be non-nil. If it is canceled or times out, ctx.Err()
// will be returned.
func (c *Client) Do(ctx context.Context, req *retryablehttp.Request, v any) error {
	// Add the context to the request.
	req = req.WithContext(ctx)

	// Execute the request and check the response.
	resp, err := c.http.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return err
		}
	}
	defer resp.Body.Close()

	// Basic response checking.
	if err := checkResponseCode(resp); err != nil {
		return err
	}

	// Return here if decoding the response isn't needed.
	if v == nil {
		return nil
	}

	// If v implements io.Writer, write the raw response body.
	if w, ok := v.(io.Writer); ok {
		_, err = io.Copy(w, resp.Body)
		return err
	}

	if err := unmarshalResponse(resp.Body, v); err != nil {
		return fmt.Errorf("unmarshalling response: %w", err)
	}
	return nil
}

func unmarshalResponse(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Get the value of model so we can test if it's a slice or struct.
	dst := reflect.Indirect(reflect.ValueOf(v))

	if dst.Kind() == reflect.Slice {
		return doUnmarshal(b, v, internal.DefaultJSONAPIUnmarshalOptions...)
	}

	// Return an error if model is not a struct, slice or an io.Writer.
	if dst.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a struct, slice or an io.Writer")
	}

	// Try to get the Items and Pagination struct fields.
	items := dst.FieldByName("Items")
	pagination := dst.FieldByName("Pagination")

	// Unmarshal a single value if v does not contain the
	// Items and Pagination struct fields.
	if !items.IsValid() || !pagination.IsValid() {
		return doUnmarshal(b, v)
	}

	// Return an error if v.Items is not a slice.
	if items.Type().Kind() != reflect.Slice {
		return fmt.Errorf("v.Items must be a slice")
	}

	err = doUnmarshal(b, items.Addr().Interface(), jsonapi.UnmarshalMeta(pagination.Addr().Interface()))
	if err != nil {
		return fmt.Errorf("unmarshalling response: %w", err)
	}

	return nil
}

func doUnmarshal(b []byte, v any, opts ...jsonapi.UnmarshalOption) error {
	opts = append(opts, internal.DefaultJSONAPIUnmarshalOptions...)
	return jsonapi.Unmarshal(b, v, opts...)
}

// checkResponseCode can be used to check the status code of an HTTP request.
func checkResponseCode(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}
	switch r.StatusCode {
	case 401:
		return internal.ErrUnauthorized
	case 404:
		return internal.ErrResourceNotFound
	case 408, 502:
		// 408 Request Timeout, 504 Gateway Timeout
		return internal.ErrTimeout
	case 409:
		return internal.ErrConflict
	}
	if err := tryUnmarshalJSONAPIError(r.Body); err != nil {
		return err
	}
	// get contents of body and log that in the error message so we know
	// what it is choking on.
	return errors.New(r.Status)
}

// tryUnmarshalJSONAPIError tries to unmarshal from the reader an error in
// JSON:API format. If it fails then it returns nil.
func tryUnmarshalJSONAPIError(r io.Reader) error {
	// Decode the error payload.
	var payload struct {
		Errors []*jsonapi.Error `json:"errors"`
	}
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil
	}
	if len(payload.Errors) == 0 {
		return nil
	}
	// Parse and format the errors.
	var errs []string
	for _, e := range payload.Errors {
		if e.Detail == "" {
			errs = append(errs, e.Title)
		} else {
			errs = append(errs, fmt.Sprintf("%s: %s", e.Title, e.Detail))
		}
	}
	return errors.New(strings.Join(errs, "\n"))
}
