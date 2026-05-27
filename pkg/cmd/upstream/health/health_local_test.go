package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubConfig struct {
	baseURL string
}

func (s *stubConfig) BaseURL() string                                 { return s.baseURL }
func (s *stubConfig) APIKey() string                                  { return "" }
func (s *stubConfig) CurrentContext() string                          { return "test" }
func (s *stubConfig) Contexts() []config.Context                      { return nil }
func (s *stubConfig) GetContext(name string) (*config.Context, error) { return nil, nil }
func (s *stubConfig) AddContext(ctx config.Context) error             { return nil }
func (s *stubConfig) RemoveContext(name string) error                 { return nil }
func (s *stubConfig) SetCurrentContext(name string) error             { return nil }
func (s *stubConfig) Save() error                                     { return nil }

func TestUpstreamHealth_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := healthRun(&Options{IO: ios})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}

func TestDeriveControlURL(t *testing.T) {
	controlURL, err := deriveControlURL("http://127.0.0.1:9180")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:9090", controlURL)
}

func TestBuildHealthURL(t *testing.T) {
	healthURL, err := buildHealthURL("http://127.0.0.1:19090", "u-1")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:19090/v1/healthcheck/upstreams/u-1", healthURL)
}

func TestBuildHealthURL_MissingHost(t *testing.T) {
	_, err := buildHealthURL("http://", "u-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing host")
}

// APISIX >= 3.x returns the per-upstream healthcheck response with
// `nodes` as a JSON object keyed by "<ip>:<port>", not an array. The
// older shape (array) must keep working for back-compat.
func TestHealthCheckResponse_UnmarshalNodesAsObject(t *testing.T) {
	body := []byte(`{"type":"http","name":"/apisix/upstreams/u-1","nodes":{"127.0.0.1:8080":{"ip":"127.0.0.1","port":8080,"status":"healthy","counter":{"success":3,"http_failure":0,"tcp_failure":0,"timeout_failure":0}}}}`)
	var resp HealthCheckResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, "http", resp.Type)
	require.Len(t, resp.Nodes, 1)
	assert.Equal(t, "127.0.0.1", resp.Nodes[0].IP)
	assert.Equal(t, 8080, resp.Nodes[0].Port)
	assert.Equal(t, "healthy", resp.Nodes[0].Status)
	assert.Equal(t, 3, resp.Nodes[0].Counter.Success)
}

func TestHealthCheckResponse_UnmarshalEmptyNodesObject(t *testing.T) {
	body := []byte(`{"type":"http","name":"/apisix/upstreams/u-1","nodes":{}}`)
	var resp HealthCheckResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Empty(t, resp.Nodes)
}

// A6_CONTROL_URL should be honored when --control-url is unset, taking
// precedence over the URL derived from the admin server (which is what
// `upstream health` falls back to for the common :9180 → :9090 case).
func TestHealthRun_HonorsA6ControlURLEnv(t *testing.T) {
	var hitPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"http","name":"/apisix/upstreams/u-1","nodes":{}}`))
	}))
	defer srv.Close()

	t.Setenv("A6_CONTROL_URL", srv.URL)

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)
	opts := &Options{
		IO: ios,
		Config: func() (config.Config, error) {
			// Admin base URL is on a totally different host:port. If env was
			// ignored and we derived from this, the test server wouldn't be hit.
			return &stubConfig{baseURL: "http://nowhere.invalid:9180"}, nil
		},
		Client:        func() (*http.Client, error) { return http.DefaultClient, nil },
		ControlClient: func() (*http.Client, error) { return http.DefaultClient, nil },
		ID:            "u-1",
		Output:        "json",
	}

	require.NoError(t, healthRun(opts))
	assert.Equal(t, "/v1/healthcheck/upstreams/u-1", hitPath, "env-var control URL should have been dialed")
	assert.True(t, strings.Contains(stdout.String(), `"type"`), "expected control response forwarded to stdout: %s", stdout.String())
}

// --control-url flag wins over A6_CONTROL_URL env var.
func TestHealthRun_FlagWinsOverEnv(t *testing.T) {
	var hitPath string
	flagSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"type":"http","name":"/apisix/upstreams/u-1","nodes":{}}`))
	}))
	defer flagSrv.Close()
	envSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("env-var URL was dialed but --control-url flag was set")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer envSrv.Close()

	t.Setenv("A6_CONTROL_URL", envSrv.URL)

	ios, _, _, _ := iostreams.Test()
	ios.SetStdoutTTY(false)
	opts := &Options{
		IO:            ios,
		Config:        func() (config.Config, error) { return &stubConfig{baseURL: "http://nowhere.invalid:9180"}, nil },
		Client:        func() (*http.Client, error) { return http.DefaultClient, nil },
		ControlClient: func() (*http.Client, error) { return http.DefaultClient, nil },
		ID:            "u-1",
		Output:        "json",
		ControlURL:    flagSrv.URL,
	}

	require.NoError(t, healthRun(opts))
	assert.Equal(t, "/v1/healthcheck/upstreams/u-1", hitPath)
}

func TestHealthCheckResponse_UnmarshalNodesAsArray(t *testing.T) {
	body := []byte(`{"type":"http","name":"/apisix/upstreams/u-1","nodes":[{"ip":"127.0.0.1","port":8080,"status":"healthy","counter":{"success":1,"http_failure":0,"tcp_failure":0,"timeout_failure":0}}]}`)
	var resp HealthCheckResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Len(t, resp.Nodes, 1)
	assert.Equal(t, "127.0.0.1", resp.Nodes[0].IP)
	assert.Equal(t, 8080, resp.Nodes[0].Port)
}
