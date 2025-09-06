package testdata

import (
	"fmt"
	"net/http"

	"github.com/petervincek/fury-flow/app"
)

// TestData holds test-specific data and dependencies for functional tests,
// including a reference to the FuryFlow application instance.
type TestData struct {
	ffApp *app.FuryFlow
}

// New creates and returns a new TestData instance initialized with the provided FuryFlow application.
func New(ffApp *app.FuryFlow) TestData {
	return TestData{
		ffApp: ffApp,
	}
}

// GetAppUrl returns the full URL of the application using the host and port
// retrieved from the ffApp instance. The URL is formatted as "http://host:port".
func (td *TestData) GetAppUrl() string {
	host := td.ffApp.GetAppHost()
	port := td.ffApp.GetAppPort()
	return fmt.Sprintf("http://%s:%d", host, port)
}

// ResponseResult is a generic struct that encapsulates the result of an HTTP response.
// It contains the raw http.Response, the HTTP status code, and a pointer to the parsed entity of type T.
type ResponseResult[T any] struct {
	response   *http.Response
	statusCode int
	entity     *T
}

// NewResponseResult creates a new ResponseResult instance containing the provided HTTP response,
// status code, and entity of generic type T.
//
// Parameters:
//
//	response   - The HTTP response associated with the result.
//	statusCode - The HTTP status code.
//	entity     - A pointer to the entity of type T returned in the response.
//
// Returns:
//
//	A ResponseResult[T] containing the response, status code, and entity.
func NewResponseResult[T any](response *http.Response, statusCode int, entity *T) ResponseResult[T] {
	return ResponseResult[T]{
		response:   response,
		statusCode: statusCode,
		entity:     entity,
	}
}

// Response returns the underlying http.Response associated with the ResponseResult.
// It provides access to the raw HTTP response for further inspection or processing.
func (r *ResponseResult[T]) Response() *http.Response {
	return r.response
}

// StatusCode returns the HTTP status code associated with the response.
// It provides access to the status code stored in the ResponseResult.
func (r *ResponseResult[T]) StatusCode() int {
	return r.statusCode
}

// Entity returns a pointer to the entity of type T contained in the ResponseResult.
// It provides access to the underlying entity value.
func (r *ResponseResult[T]) Entity() *T {
	return r.entity
}
