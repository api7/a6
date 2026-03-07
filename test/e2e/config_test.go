//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_DumpAndValidate(t *testing.T) {
	const routeID = "test-config-dump-route-1"

	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })

	createTestRoute(t, routeID, "config-dump-route", "/config-dump")

	env := setupRouteEnv(t)

	dumpFile := filepath.Join(t.TempDir(), "config.yaml")
	stdout, stderr, err := runA6WithEnv(env, "config", "dump", "-f", dumpFile)
	require.NoError(t, err, "config dump failed: stdout=%s stderr=%s", stdout, stderr)

	content, err := os.ReadFile(dumpFile)
	require.NoError(t, err)
	text := string(content)
	assert.Contains(t, text, "version: \"1\"")
	assert.Contains(t, text, "config-dump-route")
	assert.Contains(t, text, "/config-dump")

	stdout, stderr, err = runA6WithEnv(env, "config", "validate", "-f", dumpFile)
	require.NoError(t, err, "config validate failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "Config is valid")
}

func TestConfig_ValidateInvalidFile(t *testing.T) {
	env := setupRouteEnv(t)

	badFile := filepath.Join(t.TempDir(), "bad-config.yaml")
	err := os.WriteFile(badFile, []byte(`
version: "1"
routes:
  - id: "bad-route"
consumers:
  - plugins:
      key-auth:
        key: bad
`), 0o644)
	require.NoError(t, err)

	stdout, stderr, err := runA6WithEnv(env, "config", "validate", "-f", badFile)
	require.Error(t, err)
	combined := strings.ToLower(fmt.Sprintf("%s\n%s", stdout, stderr))
	assert.Contains(t, combined, "validation failed")
	assert.Contains(t, combined, "either uri or uris is required")
	assert.Contains(t, combined, "username is required")
}
