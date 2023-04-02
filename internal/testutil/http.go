package testutil

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// ExecuteRequest creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the controller
// after which the controller writes the response to the response recorder
// which we can then inspect.
func ExecuteRequest(req *http.Request, handler http.HandlerFunc) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// CheckResponseCode is a utility function to check the response code
// of the response
func CheckResponseCode(t *testing.T, expected, actual int) {
	require.Equal(t, expected, actual)
}

// CheckResponseBody is a utility function to check the response body
func CheckResponseBody(t *testing.T, body io.Reader, expected any, dst any) {
	require.NoError(t, json.NewDecoder(body).Decode(dst))
	require.Equal(t, expected, dst)
}

// GetHTTPClient returns a pre-configured HTTP client.
// #nosec
func GetHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
