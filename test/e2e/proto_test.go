//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deleteProto(t *testing.T, id string) {
	t.Helper()
	resp, err := adminAPI("DELETE", "/apisix/admin/protos/"+id, nil)
	if err == nil {
		resp.Body.Close()
	}
}

func setupProtoEnv(t *testing.T) []string {
	t.Helper()
	env := []string{"A6_CONFIG_DIR=" + t.TempDir()}
	_, _, err := runA6WithEnv(env, "context", "create", "test", "--server", adminURL, "--api-key", adminKey)
	require.NoError(t, err, "failed to create test context")
	return env
}

func TestProto_CRUD(t *testing.T) {
	const protoID = "test-proto-crud-1"

	deleteProto(t, protoID)
	t.Cleanup(func() { deleteProto(t, protoID) })

	env := setupProtoEnv(t)

	createJSON := `{
	"id": "test-proto-crud-1",
	"name": "helloworld",
	"desc": "Hello world proto for testing",
	"content": "syntax = \"proto3\";\npackage helloworld;\nservice Greeter {\n    rpc SayHello (HelloRequest) returns (HelloReply) {}\n}\nmessage HelloRequest { string name = 1; }\nmessage HelloReply { string message = 1; }"
}`
	createFile := filepath.Join(t.TempDir(), "proto.json")
	require.NoError(t, os.WriteFile(createFile, []byte(createJSON), 0o644))

	stdout, stderr, err := runA6WithEnv(env, "proto", "create", "-f", createFile)
	require.NoError(t, err, "proto create failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout+stderr, protoID)

	stdout, stderr, err = runA6WithEnv(env, "proto", "get", protoID)
	require.NoError(t, err, "proto get failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "helloworld")
	assert.Contains(t, stdout, "Hello world proto for testing")

	stdout, stderr, err = runA6WithEnv(env, "proto", "list")
	require.NoError(t, err, "proto list failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "helloworld")

	updateJSON := `{
	"name": "helloworld-updated",
	"desc": "Updated proto definition",
	"content": "syntax = \"proto3\";\npackage helloworld;\nservice Greeter {\n    rpc SayHello (HelloRequest) returns (HelloReply) {}\n    rpc SayGoodbye (HelloRequest) returns (HelloReply) {}\n}\nmessage HelloRequest { string name = 1; }\nmessage HelloReply { string message = 1; }"
}`
	updateFile := filepath.Join(t.TempDir(), "proto-updated.json")
	require.NoError(t, os.WriteFile(updateFile, []byte(updateJSON), 0o644))

	stdout, stderr, err = runA6WithEnv(env, "proto", "update", protoID, "-f", updateFile)
	require.NoError(t, err, "proto update failed: stdout=%s stderr=%s", stdout, stderr)

	stdout, stderr, err = runA6WithEnv(env, "proto", "get", protoID)
	require.NoError(t, err, "proto get after update failed: stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, "helloworld-updated")
	assert.Contains(t, stdout, "Updated proto definition")

	stdout, stderr, err = runA6WithEnv(env, "proto", "delete", protoID, "--force")
	require.NoError(t, err, "proto delete failed: stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "proto", "get", protoID)
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestProto_ListEmpty(t *testing.T) {
	env := setupProtoEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "proto", "list")
	require.NoError(t, err, "proto list failed: stdout=%s stderr=%s", stdout, stderr)
	combined := stdout + stderr
	noProtos := strings.Contains(combined, "No proto definitions found.") ||
		strings.Contains(combined, "No proto") ||
		strings.Contains(combined, "no proto") ||
		(strings.Contains(strings.ToLower(combined), "id") && !strings.Contains(combined, "helloworld"))
	assert.True(t, noProtos || strings.TrimSpace(stdout) == "[]",
		"list should indicate no proto definitions found, got: %s", combined)
}

func TestProto_GetNonExistent(t *testing.T) {
	env := setupProtoEnv(t)

	_, stderr, err := runA6WithEnv(env, "proto", "get", "nonexistent-proto-999")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestProto_DeleteNonExistent(t *testing.T) {
	env := setupProtoEnv(t)

	_, stderr, err := runA6WithEnv(env, "proto", "delete", "nonexistent-proto-999", "--force")
	assert.Error(t, err)
	assert.True(t, strings.Contains(stderr, "not found") || strings.Contains(stderr, "404"),
		"error should indicate not found, got: %s", stderr)
}

func TestProto_JSONOutput(t *testing.T) {
	const protoID = "test-proto-json-out"

	deleteProto(t, protoID)
	t.Cleanup(func() { deleteProto(t, protoID) })

	resp, err := adminAPI("PUT", "/apisix/admin/protos/"+protoID, []byte(`{
	"name":"json-output-proto",
	"desc":"proto for json output test",
	"content":"syntax = \"proto3\";\npackage jsonout;\nmessage Ping { string message = 1; }"
}`))
	require.NoError(t, err)
	resp.Body.Close()
	require.Less(t, resp.StatusCode, 400)

	env := setupProtoEnv(t)

	stdout, stderr, err := runA6WithEnv(env, "proto", "list", "--output", "json")
	require.NoError(t, err, "proto list --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "list --output json should be valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, "json-output-proto")

	stdout, stderr, err = runA6WithEnv(env, "proto", "get", protoID, "--output", "json")
	require.NoError(t, err, "proto get --output json failed: stdout=%s stderr=%s", stdout, stderr)
	assert.True(t, json.Valid([]byte(stdout)), "get --output json should be valid JSON, got: %s", stdout)
	assert.Contains(t, stdout, protoID)
	assert.Contains(t, stdout, "json-output-proto")
}
