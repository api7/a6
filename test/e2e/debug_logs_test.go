//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugLogs_DockerMode(t *testing.T) {
	env := setupRouteEnv(t)
	stdout, stderr, err := runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "10")
	require.NoError(t, err, "debug logs failed: stdout=%s stderr=%s", stdout, stderr)
	assert.NotEmpty(t, stdout)
}

func TestDebugLogs_AutoDetectContainer(t *testing.T) {
	env := setupRouteEnv(t)
	stdout, stderr, err := runA6WithEnv(env, "debug", "logs", "--tail", "10")
	if err != nil && strings.Contains(stderr, "no APISIX container found") {
		t.Skipf("skipping auto-detect test because no CI-style apisix container is running: %s", stderr)
	}
	require.NoError(t, err, "debug logs auto-detect failed: stdout=%s stderr=%s", stdout, stderr)
	assert.NotEmpty(t, stdout)
}

func TestDebugLogs_DockerModeWithSince(t *testing.T) {
	env := setupRouteEnv(t)
	stdout, stderr, err := runA6WithEnv(env, "debug", "logs", "--container", "apisix", "--tail", "5", "--since", "1h")
	require.NoError(t, err, "debug logs --since failed: stdout=%s stderr=%s", stdout, stderr)
}

func TestDebugLogs_NonExistentContainer(t *testing.T) {
	env := setupRouteEnv(t)
	_, stderr, err := runA6WithEnv(env, "debug", "logs", "--container", "non-existent-container-xyz", "--tail", "5")
	assert.Error(t, err)
	assert.Contains(t, stderr, "No such container")
}
