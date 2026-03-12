//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSkillPluginAIContentModeration verifies config-only setup of the ai-aws-content-moderation plugin.
// Data-plane testing requires AWS Comprehend credentials.
func TestSkillPluginAIContentModeration(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-ai-content-mod-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{
		"id": "skill-ai-content-mod-route",
		"uri": "/skill-ai-content-mod",
		"plugins": {
			"ai-aws-content-moderation": {
				"comprehend": {
					"endpoint": "https://comprehend.us-east-1.amazonaws.com",
					"access_key_id": "test-key-id",
					"secret_access_key": "test-secret-key",
					"region": "us-east-1"
				}
			}
		},
		"upstream": {"type": "roundrobin", "nodes": {"127.0.0.1:8080": 1}}
	}`
	f := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", f)
	if err != nil {
		t.Skipf("skipping: plugin config rejected by APISIX: stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"ai-aws-content-moderation"`)
	assert.Contains(t, stdout, `"us-east-1"`)
}
