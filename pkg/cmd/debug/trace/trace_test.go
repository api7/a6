package trace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/api7/a6/internal/config"
	"github.com/api7/a6/pkg/httpmock"
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

func TestTrace_BasicRoute(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{
		"key":"/apisix/routes/1",
		"value":{
			"id":"1",
			"name":"route-1",
			"uri":"/trace",
			"methods":["GET","POST"],
			"hosts":["example.com"],
			"plugins":{"limit-req":{},"proxy-rewrite":{}},
			"upstream_id":"ups-1"
		}
	}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/schema", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"plugins":{"limit-req":{"priority":1001},"proxy-rewrite":{"priority":1008}}}`))
	}))
	t.Cleanup(controlSrv.Close)

	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/trace", r.URL.Path)
		require.Equal(t, "example.com", r.Host)
		w.Header().Set("Apisix-Plugins", "proxy-rewrite,limit-req")
		w.Header().Set("X-APISIX-Upstream-Status", "200")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(gatewaySrv.Close)

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		ControlURL: controlSrv.URL,
		GatewayURL: gatewaySrv.URL,
	})

	require.NoError(t, err)
	out := stdout.String()
	assert.Contains(t, out, "Route:    /trace (ID: 1)")
	assert.Contains(t, out, "Request:  GET "+gatewaySrv.URL+"/trace")
	assert.Contains(t, out, "Status:   200 OK")
	assert.Contains(t, out, "Configured Plugins")
	assert.Contains(t, out, "proxy-rewrite")
	assert.Contains(t, out, "1008")
	assert.Contains(t, out, "Executed Plugins:   proxy-rewrite, limit-req")
	adminReg.Verify(t)
}

func TestTrace_NoDebugMode(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{
		"key":"/apisix/routes/1",
		"value":{"id":"1","uri":"/trace","plugins":{"ip-restriction":{}}}
	}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"plugins":{"ip-restriction":{"priority":3000}}}`))
	}))
	t.Cleanup(controlSrv.Close)

	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(gatewaySrv.Close)

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(true)

	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		ControlURL: controlSrv.URL,
		GatewayURL: gatewaySrv.URL,
	})

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "enable debug mode in APISIX")
	adminReg.Verify(t)
}

func TestTrace_JSONOutput(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{
		"key":"/apisix/routes/1",
		"value":{
			"id":"1",
			"uri":"/trace",
			"methods":["GET"],
			"plugins":{"proxy-rewrite":{}},
			"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}
		}
	}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"plugins":{"proxy-rewrite":{"priority":1008}}}`))
	}))
	t.Cleanup(controlSrv.Close)

	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Apisix-Plugins", "proxy-rewrite")
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(gatewaySrv.Close)

	ios, _, stdout, _ := iostreams.Test()
	ios.SetStdoutTTY(false)

	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		ControlURL: controlSrv.URL,
		GatewayURL: gatewaySrv.URL,
		Output:     "json",
	})

	require.NoError(t, err)
	var got TraceResult
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "1", got.Route.ID)
	assert.Equal(t, "/trace", got.Route.URI)
	assert.Equal(t, http.StatusCreated, got.Response.Status)
	assert.Equal(t, []string{"proxy-rewrite"}, got.Response.ExecutedPlugins)
	assert.Equal(t, "proxy-rewrite", got.ConfiguredPlugins[0].Name)
	adminReg.Verify(t)
}

func TestTrace_WithMethodOverride(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{"key":"/apisix/routes/1","value":{"id":"1","uri":"/trace","methods":["GET"]}}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"plugins":{}}`))
	}))
	t.Cleanup(controlSrv.Close)

	var gotMethod string
	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(gatewaySrv.Close)

	ios, _, _, _ := iostreams.Test()
	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		Method:     "POST",
		ControlURL: controlSrv.URL,
		GatewayURL: gatewaySrv.URL,
		Output:     "json",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	adminReg.Verify(t)
}

func TestTrace_WithPathOverride(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{"key":"/apisix/routes/1","value":{"id":"1","uri":"/trace"}}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"plugins":{}}`))
	}))
	t.Cleanup(controlSrv.Close)

	var gotPath string
	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(gatewaySrv.Close)

	ios, _, _, _ := iostreams.Test()
	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		Path:       "/custom",
		ControlURL: controlSrv.URL,
		GatewayURL: gatewaySrv.URL,
		Output:     "json",
	})
	require.NoError(t, err)
	assert.Equal(t, "/custom", gotPath)
	adminReg.Verify(t)
}

func TestTrace_WithHostOverride(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{"key":"/apisix/routes/1","value":{"id":"1","uri":"/trace"}}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"plugins":{}}`))
	}))
	t.Cleanup(controlSrv.Close)

	var gotHost string
	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(gatewaySrv.Close)

	ios, _, _, _ := iostreams.Test()
	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		Host:       "example.com",
		ControlURL: controlSrv.URL,
		GatewayURL: gatewaySrv.URL,
		Output:     "json",
	})
	require.NoError(t, err)
	assert.Equal(t, "example.com", gotHost)
	adminReg.Verify(t)
}

func TestTrace_RouteNotFound(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/999", httpmock.StringResponse(404, `{"error_msg":"not found"}`))

	ios, _, _, _ := iostreams.Test()
	err := traceRun(&Options{
		IO:     ios,
		Config: func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client: func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:     "999",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource not found")
	adminReg.Verify(t)
}

func TestTrace_UpstreamStatus(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{"key":"/apisix/routes/1","value":{"id":"1","uri":"/trace"}}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"plugins":{}}`))
	}))
	t.Cleanup(controlSrv.Close)

	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-APISIX-Upstream-Status", "502")
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(gatewaySrv.Close)

	ios, _, stdout, _ := iostreams.Test()
	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		ControlURL: controlSrv.URL,
		GatewayURL: gatewaySrv.URL,
		Output:     "json",
	})
	require.NoError(t, err)

	var got TraceResult
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "502", got.Response.UpstreamStatus)
	assert.Equal(t, http.StatusBadGateway, got.Response.Status)
	adminReg.Verify(t)
}

func TestTrace_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := traceRun(&Options{IO: ios})
	require.Error(t, err)
	assert.Equal(t, "route-id argument is required (or run interactively in a terminal)", err.Error())
}

func TestTrace_WithGatewayEnvOverride(t *testing.T) {
	adminReg := &httpmock.Registry{}
	adminReg.Register(http.MethodGet, "/apisix/admin/routes/1", httpmock.JSONResponse(`{"key":"/apisix/routes/1","value":{"id":"1","uri":"/trace"}}`))

	controlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"plugins":{}}`))
	}))
	t.Cleanup(controlSrv.Close)

	gatewaySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(gatewaySrv.Close)

	t.Setenv("APISIX_GATEWAY_URL", gatewaySrv.URL)

	ios, _, stdout, _ := iostreams.Test()
	err := traceRun(&Options{
		IO:         ios,
		Config:     func() (config.Config, error) { return &mockConfig{baseURL: "http://localhost:9180"}, nil },
		Client:     func() (*http.Client, error) { return adminReg.GetClient(), nil },
		ID:         "1",
		ControlURL: controlSrv.URL,
		Output:     "json",
	})
	require.NoError(t, err)

	var got TraceResult
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, fmt.Sprintf("%s/trace", gatewaySrv.URL), got.Request.URL)
	adminReg.Verify(t)
}
