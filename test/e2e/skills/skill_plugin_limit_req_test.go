//go:build e2e

package skills

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillPluginLimitReq(t *testing.T) {
	env := setupEnv(t)
	const routeID = "skill-limit-req-route"

	_, _, _ = runA6WithEnv(env, "route", "delete", routeID, "--force")
	t.Cleanup(func() { cleanupRoute(t, routeID) })

	routeJSON := `{"id":"skill-limit-req-route","uri":"/skill-limit-req","plugins":{"limit-req":{"rate":1,"burst":0,"rejected_code":429,"key_type":"var","key":"remote_addr","nodelay":true},"proxy-rewrite":{"uri":"/get"}},"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}}}`
	routeFile := writeJSON(t, "route", routeJSON)
	stdout, stderr, err := runA6WithEnv(env, "route", "create", "-f", routeFile)
	require.NoError(t, err, "route create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "route", "get", routeID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"limit-req"`)

	_, _ = httpGetWithRetry(t, gatewayURL+"/skill-limit-req", nil, 200, 5*time.Second)
	has429 := false
	for i := 0; i < 6; i++ {
		status, _ := httpGet(t, gatewayURL+"/skill-limit-req", nil)
		if status == 429 {
			has429 = true
			break
		}
	}
	assert.True(t, has429)
}
