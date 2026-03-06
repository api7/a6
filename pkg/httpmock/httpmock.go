package httpmock

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
)

// Response represents a canned HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

type mock struct {
	method string
	path   string
	resp   Response
	called int
}

// Registry is an HTTP mock registry that implements http.RoundTripper.
// Register expected request/response pairs and use GetClient() to get
// an *http.Client that uses this registry as its transport.
type Registry struct {
	mu    sync.Mutex
	mocks []mock
}

// Register adds a mock response for the given method and path.
func (r *Registry) Register(method, path string, resp Response) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mocks = append(r.mocks, mock{method: method, path: path, resp: resp})
}

// RoundTrip implements http.RoundTripper.
func (r *Registry) RoundTrip(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, m := range r.mocks {
		if m.method == req.Method && m.path == req.URL.Path {
			r.mocks[i].called++
			header := make(http.Header)
			header.Set("Content-Type", "application/json")
			for k, v := range m.resp.Header {
				header[k] = v
			}
			return &http.Response{
				StatusCode: m.resp.StatusCode,
				Body:       io.NopCloser(bytes.NewBuffer(m.resp.Body)),
				Header:     header,
			}, nil
		}
	}
	return nil, fmt.Errorf("no mock registered for %s %s", req.Method, req.URL.Path)
}

// GetClient returns an *http.Client that uses this registry as its transport.
func (r *Registry) GetClient() *http.Client {
	return &http.Client{Transport: r}
}

// Verify asserts that all registered mocks were called at least once.
func (r *Registry) Verify(t *testing.T) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.mocks {
		if m.called == 0 {
			t.Errorf("mock never called: %s %s", m.method, m.path)
		}
	}
}

// CallCount returns the number of times the mock for the given method/path was called.
func (r *Registry) CallCount(method, path string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.mocks {
		if m.method == method && m.path == path {
			return m.called
		}
	}
	return 0
}

// JSONResponse creates a Response with status 200 and the given JSON body.
func JSONResponse(body string) Response {
	return Response{
		StatusCode: http.StatusOK,
		Body:       []byte(body),
	}
}

// StringResponse creates a Response with the given status code and body.
func StringResponse(statusCode int, body string) Response {
	return Response{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
}
