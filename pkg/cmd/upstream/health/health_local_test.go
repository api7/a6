package health

import (
	"encoding/json"
	"testing"

	"github.com/api7/a6/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestHealthCheckResponse_UnmarshalNodesAsArray(t *testing.T) {
	body := []byte(`{"type":"http","name":"/apisix/upstreams/u-1","nodes":[{"ip":"127.0.0.1","port":8080,"status":"healthy","counter":{"success":1,"http_failure":0,"tcp_failure":0,"timeout_failure":0}}]}`)
	var resp HealthCheckResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Len(t, resp.Nodes, 1)
	assert.Equal(t, "127.0.0.1", resp.Nodes[0].IP)
	assert.Equal(t, 8080, resp.Nodes[0].Port)
}
