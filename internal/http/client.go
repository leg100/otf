package http

import (
	"context"
	"crypto/tls"
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
	"github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/leg100/otf/internal"
)

const (
	DefaultAddress = "localhost:8080"
)

type Client struct {
	baseURL           *url.URL
	token             string
	headers           http.Header
	http              *retryablehttp.Client
	retryLogHook      RetryLogHook
	retryServerErrors bool
	remoteAPIVersion  string
}

func NewClient(config Config) (*Client, error) {
	addr, err := SanitizeAddress(config.Address)
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

	if config.HTTPClient == nil {
		transport := cleanhttp.DefaultPooledTransport()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: config.Insecure}
		config.HTTPClient = &http.Client{Transport: transport}
	}

	// Create the client.
	client := &Client{
		baseURL:      baseURL,
		token:        config.Token,
		headers:      config.Headers,
		retryLogHook: config.RetryLogHook,
	}
	client.http = &retryablehttp.Client{
		CheckRetry:   client.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   config.HTTPClient,
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 400 * time.Millisecond,
		RetryMax:     30,
	}
	meta, err := client.getRawAPIMetadata()
	if err != nil {
		return nil, err
	}
	// Save the API version so we can return it from the RemoteAPIVersion method
	// later.
	client.remoteAPIVersion = meta.APIVersion

	return client, nil
}

// Hostname returns the server host:port.
func (c *Client) Hostname() string {
	return c.baseURL.Host
}

func (c *Client) getRawAPIMetadata() (rawAPIMetadata, error) {
	var meta rawAPIMetadata

	// Create a new request.
	u, err := c.baseURL.Parse(PingEndpoint)
	if err != nil {
		return meta, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return meta, err
	}

	// Attach the default headers.
	for k, v := range c.headers {
		req.Header[k] = v
	}
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Make a single request to retrieve the rate limit headers.
	resp, err := c.http.HTTPClient.Do(req)
	if err != nil {
		return meta, err
	}
	resp.Body.Close()

	meta.APIVersion = resp.Header.Get(headerAPIVersion)

	return meta, nil
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
	invalidBodyError := errors.New("go-tfe bug: DELETE/PATCH/POST body must be nil, ptr, or ptr slice")
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
	jsonFields := 0
	for i := 0; i < modelType.NumField(); i++ {
		structField := modelType.Field(i)
		if structField.Tag.Get("jsonapi") != "" {
			jsonAPIFields++
		}
		if structField.Tag.Get("json") != "" {
			jsonFields++
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

// RetryServerErrors configures the retry HTTP check to also retry unexpected
// errors or requests that failed with a server error.
func (c *Client) RetryServerErrors(retry bool) {
	c.retryServerErrors = retry
}

// retryHTTPCheck provides a callback for Client.CheckRetry which
// will retry both rate limit (429) and server (>= 500) errors.
func (c *Client) retryHTTPCheck(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return c.retryServerErrors, err
	}
	if resp.StatusCode == 429 || (c.retryServerErrors && resp.StatusCode >= 500) {
		return true, nil
	}
	return false, nil
}

func unmarshalResponse(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Get the value of model so we can test if it's a struct.
	dst := reflect.Indirect(reflect.ValueOf(v))

	// Return an error if model is not a struct or an io.Writer.
	if dst.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a struct or an io.Writer")
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
	}
	// Decode the error payload.
	var payload struct {
		Errors []*jsonapi.Error `json:"errors"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return fmt.Errorf("unable to decode errors payload: %w", err)
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
