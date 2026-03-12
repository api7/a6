//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginSkywalking verifies config-only setup of the skywalking plugin.
// Data-plane testing requires a SkyWalking OAP endpoint.
func TestSkillPluginSkywalking(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-skywalking-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-skywalking-route",
		"uri": "/skill-skywalking",
		"plugins": {
			"skywalking": {
				"sample_ratio": 1
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	if err != nil {
		t.Skipf("skipping: plugin not available: stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"skywalking"`)
	assert.Contains(t, stdout, `"sample_ratio"`)
}
