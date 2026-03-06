//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmoke_BinaryRuns(t *testing.T) {
	stdout, _, err := runA6("--help")
	require.NoError(t, err, "a6 --help should exit successfully")
	assert.Contains(t, stdout, "a6", "help output should mention a6")
	assert.Contains(t, stdout, "Apache APISIX", "help output should mention Apache APISIX")
}

func TestSmoke_APISIXReachable(t *testing.T) {
	resp, err := adminAPI("GET", "/apisix/admin/routes", nil)
	require.NoError(t, err, "should be able to reach APISIX Admin API")
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode, "Admin API should return 200")
}

func TestSmoke_GatewayReachable(t *testing.T) {
	resp, err := adminAPI("GET", "/apisix/admin/services", nil)
	require.NoError(t, err, "should be able to reach APISIX services endpoint")
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode, "services endpoint should return 200")
}
