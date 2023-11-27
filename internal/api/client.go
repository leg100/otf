package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/DataDog/jsonapi"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
)

type (
	Client struct {
		baseURL *url.URL
		token   string
		headers http.Header
		http    *retryablehttp.Client
	}

	// Config provides configuration details to the API client.
	Config struct {
		// The address of the otf API.
		Address string
		// The base path on which the API is served.
		BasePath string
		// API token used to access the otf API.
		Token string
		// Headers that will be added to every request.
		Headers http.Header
		// Toggle retrying requests upon encountering transient errors.
		RetryRequests bool
		// RetryLogHook is invoked each time a request is retried.
		RetryLogHook retryablehttp.RequestLogHook
		// Override default http transport
		Transport http.RoundTripper
	}
)

func NewClient(config Config) (*Client, error) {
	// set defaults
	if config.Address == "" {
		config.Address = DefaultAddress
	}
	if config.BasePath == "" {
		config.BasePath = DefaultBasePath
	}
	if config.Headers == nil {
		config.Headers = make(http.Header)
	}
	if config.Transport == nil {
		config.Transport = http.DefaultTransport
	}
	config.Headers.Set("User-Agent", "otf-agent")

	addr, err := otfhttp.SanitizeAddress(config.Address)
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}
	baseURL.Path = config.BasePath
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
		token:   config.Token,
		headers: config.Headers,
	}
	client.http = &retryablehttp.Client{
		Backoff:        retryablehttp.DefaultBackoff,
		ErrorHandler:   retryablehttp.PassthroughErrorHandler,
		RequestLogHook: config.RetryLogHook,
		HTTPClient:     &http.Client{Transport: config.Transport},
		RetryWaitMin:   100 * time.Millisecond,
		RetryWaitMax:   400 * time.Millisecond,
		RetryMax:       30,
	}
	if config.RetryRequests {
		// enable retries
		client.http.CheckRetry = retryablehttp.DefaultRetryPolicy
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

// SetHeader sets a http header to be used on all requests. Not concurrent-safe.
func (c *Client) SetHeader(k, v string) {
	c.headers[k] = []string{v}
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
func (c *Client) NewRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Authorization", "Bearer "+c.token)

	var body interface{}
	switch method {
	case "GET":
		reqHeaders.Set("Accept", "application/vnd.api+json")

		if v != nil {
			q := url.Values{}
			if err := otfhttp.Encoder.Encode(v, q); err != nil {
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
	for k, v := range c.headers {
		req.Header[k] = v
	}

	// Set the request specific headers.
	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	return req, nil
}

// Helper method that serializes the given ptr or ptr slice into a JSON
// request. It automatically uses jsonapi or json serialization, depending
// on the body type's tags.
func serializeRequestBody(v interface{}) (interface{}, error) {
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
		if sliceElem.Kind() != reflect.Ptr {
			return nil, invalidBodyError
		}
		modelType = sliceElem.Elem()
	case reflect.Ptr:
		modelType = reflect.ValueOf(v).Elem().Type()
	default:
		return nil, invalidBodyError
	}

	// Infer whether the request uses jsonapi or regular json
	// serialization based on how the fields are tagged.
	jsonAPIFields := 0
	for i := 0; i < modelType.NumField(); i++ {
		structField := modelType.Field(i)
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
func (c *Client) Do(ctx context.Context, req *retryablehttp.Request, v interface{}) error {
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

	return unmarshalResponse(resp.Body, v)
}

func unmarshalResponse(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Get the value of model so we can test if it's a slice or struct.
	dst := reflect.Indirect(reflect.ValueOf(v))

	if dst.Kind() == reflect.Slice {
		return jsonapi.Unmarshal(b, v)
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
		return jsonapi.Unmarshal(b, v)
	}

	// Return an error if v.Items is not a slice.
	if items.Type().Kind() != reflect.Slice {
		return fmt.Errorf("v.Items must be a slice")
	}

	err = jsonapi.Unmarshal(b, items.Addr().Interface(), jsonapi.UnmarshalMeta(pagination.Addr().Interface()))
	if err != nil {
		return err
	}

	return nil
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
	// get contents of body and log that in the error message so we know
	// what it is choking on.
	contents, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	// Decode the error payload.
	var payload struct {
		Errors []*jsonapi.Error `json:"errors"`
	}
	if err := json.Unmarshal(contents, &payload); err != nil {
		return fmt.Errorf("unable to decode errors payload: %s: %w", string(contents), err)
	}
	if len(payload.Errors) == 0 {
		return fmt.Errorf(r.Status)
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
	return fmt.Errorf(strings.Join(errs, "\n"))
}
