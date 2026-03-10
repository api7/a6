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

func TestConsumerExport_BasicYAML(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/consumers" {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"total":1,"list":[{"key":"/apisix/consumers/c1","value":{"username":"jack","desc":"test consumer","create_time":1,"update_time":2}}]}`)),
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
	assert.Contains(t, out, "username: jack")
	assert.Contains(t, out, "desc: test consumer")
	assert.NotContains(t, out, "create_time")
	assert.NotContains(t, out, "update_time")
}

func TestConsumerExport_WithLabelFilter(t *testing.T) {
	calledWithLabel := false
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/consumers" {
			if req.URL.Query().Get("label") == "env" {
				calledWithLabel = true
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"total":1,"list":[{"key":"/apisix/consumers/c1","value":{"username":"jack","labels":{"env":"test"}}}]}`)),
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
	c.SetArgs([]string{"--label", "env", "--output", "json"})
	err := c.Execute()

	require.NoError(t, err)
	assert.True(t, calledWithLabel, "should send label to API")
	out := stdout.String()
	assert.Contains(t, out, "jack")
}

func TestConsumerExport_EmptyResult(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/consumers" {
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
	assert.Contains(t, stderr.String(), "No consumers found.")
}

func TestConsumerExport_APIError(t *testing.T) {
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

func TestConsumerExport_CorrectAPIPath(t *testing.T) {
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

	assert.Equal(t, "/apisix/admin/consumers", requestedPath)
}
