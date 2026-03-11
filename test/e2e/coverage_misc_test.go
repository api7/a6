//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionUpgradeAllEmpty(t *testing.T) {
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	stdout, stderr, err := runA6WithEnv(env, "extension", "upgrade", "--all")

	combined := stdout + stderr
	assert.True(t,
		err == nil || strings.Contains(combined, "No extensions") || strings.Contains(combined, "no extensions"),
		"extension upgrade --all should handle empty state gracefully, got: stdout=%s stderr=%s err=%v", stdout, stderr, err,
	)
}

func TestConsumerGroup_DeleteNonExistent(t *testing.T) {
	env := setupConsumerGroupEnv(t)

	_, stderr, err := runA6WithEnv(env, "consumer-group", "delete", "nonexistent-cg-999", "--force")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestPluginConfig_DeleteNonExistent(t *testing.T) {
	env := setupPluginConfigEnv(t)

	_, stderr, err := runA6WithEnv(env, "plugin-config", "delete", "nonexistent-pc-999", "--force")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestUpstream_DeleteAll(t *testing.T) {
	const (
		id1 = "misc-up-all-1"
		id2 = "misc-up-all-2"
	)

	env := setupUpstreamEnv(t)

	deleteUpstreamViaCLI(t, env, id1)
	deleteUpstreamViaCLI(t, env, id2)
	t.Cleanup(func() {
		deleteUpstreamViaAdmin(t, id1)
		deleteUpstreamViaAdmin(t, id2)
	})

	createTestUpstreamViaCLI(t, env, id1, "misc-upstream-all-one")
	createTestUpstreamViaCLI(t, env, id2, "misc-upstream-all-two")

	_, _, _ = runA6WithEnv(env, "route", "delete", "test-route", "--force")

	stdout, stderr, err := runA6WithEnv(env, "upstream", "delete", "--all", "--force")
	require.NoError(t, err, "upstream delete --all failed: stdout=%s stderr=%s", stdout, stderr)

	resp1, err := adminAPI(http.MethodGet, "/apisix/admin/upstreams/"+id1, nil)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp1.StatusCode)

	resp2, err := adminAPI(http.MethodGet, "/apisix/admin/upstreams/"+id2, nil)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestService_DeleteAll(t *testing.T) {
	const (
		id1 = "misc-svc-all-1"
		id2 = "misc-svc-all-2"
	)

	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, id1)
	deleteServiceViaCLI(t, env, id2)
	t.Cleanup(func() {
		deleteServiceViaAdmin(t, id1)
		deleteServiceViaAdmin(t, id2)
	})

	createTestServiceViaCLI(t, env, id1, "misc-service-all-one")
	createTestServiceViaCLI(t, env, id2, "misc-service-all-two")

	stdout, stderr, err := runA6WithEnv(env, "service", "delete", "--all", "--force")
	require.NoError(t, err, "service delete --all failed: stdout=%s stderr=%s", stdout, stderr)

	resp1, err := adminAPI(http.MethodGet, "/apisix/admin/services/"+id1, nil)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp1.StatusCode)

	resp2, err := adminAPI(http.MethodGet, "/apisix/admin/services/"+id2, nil)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestConsumer_DeleteAll(t *testing.T) {
	const (
		username1 = "misc-cons-all-1"
		username2 = "misc-cons-all-2"
	)

	env := setupConsumerEnv(t)

	deleteConsumerViaCLI(t, env, username1)
	deleteConsumerViaCLI(t, env, username2)
	t.Cleanup(func() {
		deleteConsumerViaAdmin(t, username1)
		deleteConsumerViaAdmin(t, username2)
	})

	createConsumerViaCLI(t, env, username1, `{"username":"misc-cons-all-1","plugins":{"key-auth":{"key":"misc-cons-all-key-1"}}}`)
	createConsumerViaCLI(t, env, username2, `{"username":"misc-cons-all-2","plugins":{"key-auth":{"key":"misc-cons-all-key-2"}}}`)

	stdout, stderr, err := runA6WithEnv(env, "consumer", "delete", "--all", "--force")
	require.NoError(t, err, "consumer delete --all failed: stdout=%s stderr=%s", stdout, stderr)

	resp1, err := adminAPI(http.MethodGet, "/apisix/admin/consumers/"+username1, nil)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp1.StatusCode)

	resp2, err := adminAPI(http.MethodGet, "/apisix/admin/consumers/"+username2, nil)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestUpstream_DeleteByLabel(t *testing.T) {
	const (
		id1 = "misc-up-lbl-1"
		id2 = "misc-up-lbl-2"
		id3 = "misc-up-lbl-prod"
	)

	env := setupUpstreamEnv(t)

	deleteUpstreamViaCLI(t, env, id1)
	deleteUpstreamViaCLI(t, env, id2)
	deleteUpstreamViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteUpstreamViaAdmin(t, id1)
		deleteUpstreamViaAdmin(t, id2)
		deleteUpstreamViaAdmin(t, id3)
	})

	body1 := `{"id":"misc-up-lbl-1","name":"labeled-up-1","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"test"}}`
	body2 := `{"id":"misc-up-lbl-2","name":"labeled-up-2","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"test"}}`
	body3 := `{"id":"misc-up-lbl-prod","name":"labeled-up-prod","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"prod"}}`

	file1 := filepath.Join(t.TempDir(), "upstream-label-1.json")
	file2 := filepath.Join(t.TempDir(), "upstream-label-2.json")
	file3 := filepath.Join(t.TempDir(), "upstream-label-3.json")
	require.NoError(t, os.WriteFile(file1, []byte(body1), 0o644))
	require.NoError(t, os.WriteFile(file2, []byte(body2), 0o644))
	require.NoError(t, os.WriteFile(file3, []byte(body3), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", file1)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)
	stdout, stderr, err = runA6WithEnv(env, "upstream", "create", "-f", file2)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)
	stdout, stderr, err = runA6WithEnv(env, "upstream", "create", "-f", file3)
	require.NoError(t, err, "upstream create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "upstream", "delete", "--label", "env=test", "--force")
	require.NoError(t, err, "upstream delete --label failed: stdout=%s stderr=%s", stdout, stderr)

	resp1, err := adminAPI(http.MethodGet, "/apisix/admin/upstreams/"+id1, nil)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp1.StatusCode)

	resp2, err := adminAPI(http.MethodGet, "/apisix/admin/upstreams/"+id2, nil)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)

	resp3, err := adminAPI(http.MethodGet, "/apisix/admin/upstreams/"+id3, nil)
	require.NoError(t, err)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusOK, resp3.StatusCode)
}

func TestService_DeleteByLabel(t *testing.T) {
	const (
		id1 = "misc-svc-lbl-1"
		id2 = "misc-svc-lbl-2"
		id3 = "misc-svc-lbl-prod"
	)

	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, id1)
	deleteServiceViaCLI(t, env, id2)
	deleteServiceViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteServiceViaAdmin(t, id1)
		deleteServiceViaAdmin(t, id2)
		deleteServiceViaAdmin(t, id3)
	})

	body1 := `{"id":"misc-svc-lbl-1","name":"labeled-svc-1","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"test"}}`
	body2 := `{"id":"misc-svc-lbl-2","name":"labeled-svc-2","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"test"}}`
	body3 := `{"id":"misc-svc-lbl-prod","name":"labeled-svc-prod","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"prod"}}`

	file1 := filepath.Join(t.TempDir(), "service-label-1.json")
	file2 := filepath.Join(t.TempDir(), "service-label-2.json")
	file3 := filepath.Join(t.TempDir(), "service-label-3.json")
	require.NoError(t, os.WriteFile(file1, []byte(body1), 0o644))
	require.NoError(t, os.WriteFile(file2, []byte(body2), 0o644))
	require.NoError(t, os.WriteFile(file3, []byte(body3), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", file1)
	require.NoError(t, err, "service create failed: stdout=%s stderr=%s", stdout, stderr)
	stdout, stderr, err = runA6WithEnv(env, "service", "create", "-f", file2)
	require.NoError(t, err, "service create failed: stdout=%s stderr=%s", stdout, stderr)
	stdout, stderr, err = runA6WithEnv(env, "service", "create", "-f", file3)
	require.NoError(t, err, "service create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "service", "delete", "--label", "env=test", "--force")
	require.NoError(t, err, "service delete --label failed: stdout=%s stderr=%s", stdout, stderr)

	resp1, err := adminAPI(http.MethodGet, "/apisix/admin/services/"+id1, nil)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp1.StatusCode)

	resp2, err := adminAPI(http.MethodGet, "/apisix/admin/services/"+id2, nil)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)

	resp3, err := adminAPI(http.MethodGet, "/apisix/admin/services/"+id3, nil)
	require.NoError(t, err)
	defer resp3.Body.Close()
	assert.Equal(t, http.StatusOK, resp3.StatusCode)
}

func TestRoute_ExportToFileByLabel(t *testing.T) {
	const (
		id1 = "misc-route-exp-file-1"
		id2 = "misc-route-exp-file-2"
		id3 = "misc-route-exp-file-prod"
	)

	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, id1)
	deleteRouteViaCLI(t, env, id2)
	deleteRouteViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, id1)
		deleteRouteViaAdmin(t, id2)
		deleteRouteViaAdmin(t, id3)
	})

	createLabeledRouteViaCLI(t, env, id1, "/misc/export/file/1", "env", "test")
	createLabeledRouteViaCLI(t, env, id2, "/misc/export/file/2", "env", "test")
	createLabeledRouteViaCLI(t, env, id3, "/misc/export/file/3", "env", "prod")

	outFile := filepath.Join(t.TempDir(), "export-routes.yaml")
	stdout, stderr, err := runA6WithEnv(env, "route", "export", "--label", "env=test", "-f", outFile)
	require.NoError(t, err, "route export -f failed: stdout=%s stderr=%s", stdout, stderr)

	contentBytes, err := os.ReadFile(outFile)
	require.NoError(t, err)
	content := string(contentBytes)

	assert.Contains(t, content, id1)
	assert.Contains(t, content, id2)
	assert.NotContains(t, content, id3)
	assert.Contains(t, content, "uri:")
}

func TestRoute_ExportYAMLOutput(t *testing.T) {
	const id = "misc-route-export-yaml-1"

	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, id)
	t.Cleanup(func() { deleteRouteViaAdmin(t, id) })

	routeJSON := fmt.Sprintf(`{"id":"%s","uri":"/misc/export/yaml","name":"misc-route-export-yaml","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"suite":"misc-export-yaml"}}`, id)
	routeFile := filepath.Join(t.TempDir(), "route-export-yaml.json")
	require.NoError(t, os.WriteFile(routeFile, []byte(routeJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "route", "export", "--label", "suite=misc-export-yaml", "--output", "yaml")
	require.NoError(t, err, "route export --output yaml failed: stdout=%s stderr=%s", stdout, stderr)

	trimmed := strings.TrimSpace(stdout)
	assert.Contains(t, stdout, "id:")
	assert.Contains(t, stdout, "uri:")
	assert.Contains(t, stdout, "name:")
	assert.Contains(t, stdout, id)
	assert.False(t, strings.HasPrefix(trimmed, "{"), "yaml output should not start with JSON object")
	assert.False(t, strings.HasPrefix(trimmed, "["), "yaml output should not start with JSON array")
}
