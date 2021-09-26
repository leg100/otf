package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"golang.org/x/time/rate"
)

const (
	DefaultAddress = "localhost:8080"
)

type Client interface {
	Organizations() otf.OrganizationService
	Workspaces() otf.WorkspaceService
}

type client struct {
	baseURL           *url.URL
	token             string
	headers           http.Header
	http              *retryablehttp.Client
	limiter           *rate.Limiter
	retryLogHook      RetryLogHook
	retryServerErrors bool
	remoteAPIVersion  string

	ConfigurationVersionService otf.ConfigurationVersionService
	EventService                otf.EventService
	OrganizationService         otf.OrganizationService
	RunService                  otf.RunService
	StateVersionService         otf.StateVersionService
	WorkspaceService            otf.WorkspaceService
}

func (c *client) Organizations() otf.OrganizationService {
	return c.OrganizationService
}

func (c *client) Workspaces() otf.WorkspaceService {
	return c.WorkspaceService
}

func (c *client) getRawAPIMetadata() (rawAPIMetadata, error) {
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
	meta.RateLimit = resp.Header.Get(headerRateLimit)

	return meta, nil
}

// configureLimiter configures the rate limiter.
func (c *client) configureLimiter(rawLimit string) {

	// Set default values for when rate limiting is disabled.
	limit := rate.Inf
	burst := 0

	if v := rawLimit; v != "" {
		if rateLimit, _ := strconv.ParseFloat(v, 64); rateLimit > 0 {
			// Configure the limit and burst using a split of 2/3 for the limit and
			// 1/3 for the burst. This enables clients to burst 1/3 of the allowed
			// calls before the limiter kicks in. The remaining calls will then be
			// spread out evenly using intervals of time.Second / limit which should
			// prevent hitting the rate limit.
			limit = rate.Limit(rateLimit * 0.66)
			burst = int(rateLimit * 0.33)
		}
	}

	// Create a new limiter using the calculated values.
	c.limiter = rate.NewLimiter(limit, burst)
}

// newRequest creates an API request with proper headers and serialization.
//
// A relative URL path can be provided, in which case it is resolved relative to the baseURL
// of the Client. Relative URL paths should always be specified without a preceding slash. Adding a
// preceding slash allows for ignoring the configured baseURL for non-standard endpoints.
//
// If v is supplied, the value will be JSONAPI encoded and included as the
// request body. If the method is GET, the value will be parsed and added as
// query parameters.
func (c *client) newRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
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
			if err := encoder.Encode(v, q); err != nil {
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
	if jsonAPIFields > 0 && jsonFields > 0 {
		// Defining a struct with both json and jsonapi tags doesn't
		// make sense, because a struct can only be serialized
		// as one or another. If this does happen, it's a bug
		// in the library that should be fixed at development time
		return nil, errors.New("go-tfe bug: struct can't use both json and jsonapi attributes")
	}

	if jsonFields > 0 {
		return json.Marshal(v)
	} else {
		buf := bytes.NewBuffer(nil)
		if err := jsonapi.MarshalPayloadWithoutIncluded(buf, v); err != nil {
			return nil, err
		}
		return buf, nil
	}
}

// do sends an API request and returns the API response. The API response
// is JSONAPI decoded and the document's primary data is stored in the value
// pointed to by v, or returned as an error if an API error has occurred.

// If v implements the io.Writer interface, the raw response body will be
// written to v, without attempting to first decode it.
//
// The provided ctx must be non-nil. If it is canceled or times out, ctx.Err()
// will be returned.
func (c *client) do(ctx context.Context, req *retryablehttp.Request, v interface{}) error {
	// Wait will block until the limiter can obtain a new token
	// or returns an error if the given context is canceled.
	if err := c.limiter.Wait(ctx); err != nil {
		return err
	}

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
func (c *client) RetryServerErrors(retry bool) {
	c.retryServerErrors = retry
}

// retryHTTPCheck provides a callback for Client.CheckRetry which
// will retry both rate limit (429) and server (>= 500) errors.
func (c *client) retryHTTPCheck(ctx context.Context, resp *http.Response, err error) (bool, error) {
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

// retryHTTPBackoff provides a generic callback for Client.Backoff which
// will pass through all calls based on the status code of the response.
func (c *client) retryHTTPBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if c.retryLogHook != nil {
		c.retryLogHook(attemptNum, resp)
	}

	// Use the rate limit backoff function when we are rate limited.
	if resp != nil && resp.StatusCode == 429 {
		return rateLimitBackoff(min, max, attemptNum, resp)
	}

	// Set custom duration's when we experience a service interruption.
	min = 700 * time.Millisecond
	max = 900 * time.Millisecond

	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
}

// rateLimitBackoff provides a callback for Client.Backoff which will use the
// X-RateLimit_Reset header to determine the time to wait. We add some jitter
// to prevent a thundering herd.
//
// min and max are mainly used for bounding the jitter that will be added to
// the reset time retrieved from the headers. But if the final wait time is
// less then min, min will be used instead.
func rateLimitBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// rnd is used to generate pseudo-random numbers.
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// First create some jitter bounded by the min and max durations.
	jitter := time.Duration(rnd.Float64() * float64(max-min))

	if resp != nil {
		if v := resp.Header.Get(headerRateReset); v != "" {
			if reset, _ := strconv.ParseFloat(v, 64); reset > 0 {
				// Only update min if the given time to wait is longer.
				if wait := time.Duration(reset * 1e9); wait > min {
					min = wait
				}
			}
		}
	}

	return min + jitter
}

func unmarshalResponse(responseBody io.Reader, model interface{}) error {
	// Get the value of model so we can test if it's a struct.
	dst := reflect.Indirect(reflect.ValueOf(model))

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
		return jsonapi.UnmarshalPayload(responseBody, model)
	}

	// Return an error if v.Items is not a slice.
	if items.Type().Kind() != reflect.Slice {
		return fmt.Errorf("v.Items must be a slice")
	}

	// Create a temporary buffer and copy all the read data into it.
	body := bytes.NewBuffer(nil)
	reader := io.TeeReader(responseBody, body)

	// Unmarshal as a list of values as v.Items is a slice.
	raw, err := jsonapi.UnmarshalManyPayload(reader, items.Type().Elem())
	if err != nil {
		return err
	}

	// Make a new slice to hold the results.
	sliceType := reflect.SliceOf(items.Type().Elem())
	result := reflect.MakeSlice(sliceType, 0, len(raw))

	// Add all of the results to the new slice.
	for _, v := range raw {
		result = reflect.Append(result, reflect.ValueOf(v))
	}

	// Pointer-swap the result.
	items.Set(result)

	// As we are getting a list of values, we need to decode
	// the pagination details out of the response body.
	p, err := parsePagination(body)
	if err != nil {
		return err
	}

	// Pointer-swap the decoded pagination details.
	pagination.Set(reflect.ValueOf(p))

	return nil
}
