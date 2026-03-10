package export

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

type mockConfig struct {
	baseURL string
}

func (m *mockConfig) BaseURL() string                                 { return m.baseURL }
func (m *mockConfig) APIKey() string                                  { return "" }
func (m *mockConfig) CurrentContext() string                          { return "test" }
func (m *mockConfig) Contexts() []config.Context                      { return nil }
func (m *mockConfig) GetContext(name string) (*config.Context, error) { return nil, nil }
func (m *mockConfig) AddContext(ctx config.Context) error             { return nil }
func (m *mockConfig) RemoveContext(name string) error                 { return nil }
func (m *mockConfig) SetCurrentContext(name string) error             { return nil }
func (m *mockConfig) Save() error                                     { return nil }

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestStreamRouteExport_BasicYAML(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/stream_routes" {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"total":1,"list":[{"key":"/apisix/stream_routes/sr1","value":{"id":"sr1","server_port":9100,"remote_addr":"10.0.0.0/8","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:1980":1}},"create_time":1,"update_time":2}}]}`)),
			}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"error_msg":"not found"}`))}, nil
	})

	ios, _, stdout, _ := iostreams.Test()
	f := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: transport}, nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdExport(f)
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	out := stdout.String()
	assert.Contains(t, out, "id: sr1")
	assert.Contains(t, out, "server_port: 9100")
	assert.Contains(t, out, "roundrobin")
	assert.NotContains(t, out, "create_time")
	assert.NotContains(t, out, "update_time")
}

func TestStreamRouteExport_WithLabelFilter(t *testing.T) {
	calledWithLabel := ""
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/stream_routes" {
			calledWithLabel = req.URL.Query().Get("label")
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"total":1,"list":[{"key":"/apisix/stream_routes/sr1","value":{"id":"sr1","server_port":9100}}]}`)),
			}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"error_msg":"not found"}`))}, nil
	})

	ios, _, stdout, _ := iostreams.Test()
	f := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: transport}, nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdExport(f)
	c.SetArgs([]string{"--label", "env=staging", "--output", "json"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Equal(t, "env=staging", calledWithLabel, "should pass full label to API query")
	out := stdout.String()
	assert.Contains(t, out, "sr1")
}

func TestStreamRouteExport_EmptyResult(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/stream_routes" {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"total":0,"list":{}}`)),
			}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"error_msg":"not found"}`))}, nil
	})

	ios, _, stdout, stderr := iostreams.Test()
	f := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: transport}, nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdExport(f)
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.NoError(t, err)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "No stream routes found.")
}

func TestStreamRouteExport_APIError(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"error_msg":"internal error"}`)),
		}, nil
	})

	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: transport}, nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdExport(f)
	c.SetArgs([]string{"--output", "yaml"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal error")
}

func TestStreamRouteExport_CorrectAPIPath(t *testing.T) {
	var requestedPath string
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requestedPath = req.URL.Path
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"total":0,"list":{}}`)),
		}, nil
	})

	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: transport}, nil
		},
		Config: func() (config.Config, error) {
			return &mockConfig{baseURL: "http://localhost:9180"}, nil
		},
	}

	c := NewCmdExport(f)
	c.SetArgs([]string{"--output", "yaml"})
	_ = c.Execute()

	assert.Equal(t, "/apisix/admin/stream_routes", requestedPath)
}
