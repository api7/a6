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

func TestRouteExport_BasicYAML(t *testing.T) {
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/routes" {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"total":1,"list":[{"key":"/apisix/routes/r1","value":{"id":"r1","name":"r-1","uri":"/r1","create_time":1,"update_time":2}}]}`)),
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
	assert.Contains(t, out, "id: r1")
	assert.Contains(t, out, "name: r-1")
	assert.NotContains(t, out, "create_time")
	assert.NotContains(t, out, "update_time")
}

func TestRouteExport_WithLabelFilter(t *testing.T) {
	calledWithNormalizedLabel := false
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && req.URL.Path == "/apisix/admin/routes" {
			if req.URL.Query().Get("label") == "env:test" {
				calledWithNormalizedLabel = true
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"total":1,"list":[{"key":"/apisix/routes/r1","value":{"id":"r1","name":"match","uri":"/r1","labels":{"env":"test"}}}]}`)),
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
	c.SetArgs([]string{"--label", "env=test", "--output", "json"})
	err := c.Execute()

	require.NoError(t, err)
	assert.True(t, calledWithNormalizedLabel, "should send normalized label key:value to API")
	out := stdout.String()
	assert.Contains(t, out, "match")
}
