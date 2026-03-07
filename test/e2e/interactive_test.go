//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInteractive_RouteGetRequiresIDInNonTTY(t *testing.T) {
	_, stderr, err := runA6("route", "get")
	require.Error(t, err)
	assert.Contains(t, stderr, "id argument is required")
}

func TestInteractive_UpstreamHealthRequiresIDInNonTTY(t *testing.T) {
	_, stderr, err := runA6("upstream", "health")
	require.Error(t, err)
	assert.Contains(t, stderr, "id argument is required")
}

func TestInteractive_ExplicitIDStillWorks(t *testing.T) {
	const routeID = "test-interactive-explicit-id"
	deleteRoute(t, routeID)
	t.Cleanup(func() { deleteRoute(t, routeID) })
	createTestRoute(t, routeID, "interactive-explicit", "/interactive-explicit")

	env := setupRouteEnv(t)
	stdout, stderr, err := runA6WithEnv(env, "route", "get", routeID)
	require.NoError(t, err, "route get with explicit ID failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "interactive-explicit")
}
