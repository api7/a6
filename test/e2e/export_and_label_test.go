//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeJSONFile(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

func parseJSONArrayOutput(t *testing.T, stdout string) []map[string]interface{} {
	t.Helper()

	var list []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &list); err == nil {
		return list
	}

	var wrapped map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &wrapped), "stdout=%s", stdout)
	if nodes, ok := wrapped["list"].([]interface{}); ok {
		list = make([]map[string]interface{}, 0, len(nodes))
		for _, n := range nodes {
			if m, ok := n.(map[string]interface{}); ok {
				list = append(list, m)
			}
		}
	}
	return list
}

func collectFieldValues(items []map[string]interface{}, field string) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		if v, ok := item[field].(string); ok {
			values = append(values, v)
		}
	}
	return values
}

func deleteProtoViaCLI(t *testing.T, env []string, id string) {
	t.Helper()
	_, _, _ = runA6WithEnv(env, "proto", "delete", id, "--force")
}

func deleteSSLViaCLI(t *testing.T, env []string, id string) {
	t.Helper()
	_, _, _ = runA6WithEnv(env, "ssl", "delete", id, "--force")
}

func deleteStreamRouteViaCLI(t *testing.T, env []string, id string) {
	t.Helper()
	_, _, _ = runA6WithEnv(env, "stream-route", "delete", id, "--force")
}

func TestExport_Upstream(t *testing.T) {
	id1 := "exp-up-1"
	id2 := "exp-up-2"
	id3 := "exp-up-3"
	env := setupUpstreamEnv(t)

	deleteUpstreamViaCLI(t, env, id1)
	deleteUpstreamViaCLI(t, env, id2)
	deleteUpstreamViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteUpstreamViaAdmin(t, id1)
		deleteUpstreamViaAdmin(t, id2)
		deleteUpstreamViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"suite":"exp-up"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"suite":"exp-up"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"suite":"exp-up-other"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("upstream-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "upstream", "export", "--label", "suite=exp-up", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	items := parseJSONArrayOutput(t, stdout)
	ids := strings.Join(collectFieldValues(items, "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "upstream", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "upstream-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "upstream", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_Service(t *testing.T) {
	id1 := "exp-svc-1"
	id2 := "exp-svc-2"
	id3 := "exp-svc-3"
	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, id1)
	deleteServiceViaCLI(t, env, id2)
	deleteServiceViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteServiceViaAdmin(t, id1)
		deleteServiceViaAdmin(t, id2)
		deleteServiceViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"suite":"exp-svc"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"suite":"exp-svc"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"suite":"exp-svc-other"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("service-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "service", "export", "--label", "suite=exp-svc", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "service", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "service-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "service", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_Consumer(t *testing.T) {
	id1 := "exp-consumer-1"
	id2 := "exp-consumer-2"
	id3 := "exp-consumer-3"
	env := setupConsumerEnv(t)

	deleteConsumerViaCLI(t, env, id1)
	deleteConsumerViaCLI(t, env, id2)
	deleteConsumerViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteConsumerViaAdmin(t, id1)
		deleteConsumerViaAdmin(t, id2)
		deleteConsumerViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"username":"%s","plugins":{"key-auth":{"key":"%s-key"}},"labels":{"suite":"exp-consumer"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"username":"%s","plugins":{"key-auth":{"key":"%s-key"}},"labels":{"suite":"exp-consumer"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"username":"%s","plugins":{"key-auth":{"key":"%s-key"}},"labels":{"suite":"exp-consumer-other"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("consumer-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "consumer", "export", "--label", "suite=exp-consumer", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	users := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "username"), ",")
	assert.Contains(t, users, id1)
	assert.Contains(t, users, id2)
	assert.NotContains(t, users, id3)

	stdout, stderr, err = runA6WithEnv(env, "consumer", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allUsers := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "username"), ",")
	assert.Contains(t, allUsers, id1)
	assert.Contains(t, allUsers, id2)
	assert.Contains(t, allUsers, id3)

	outFile := filepath.Join(t.TempDir(), "consumer-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "consumer", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_ConsumerGroup(t *testing.T) {
	id1 := "exp-cg-1"
	id2 := "exp-cg-2"
	id3 := "exp-cg-3"
	env := setupConsumerGroupEnv(t)

	deleteConsumerGroupViaCLI(t, env, id1)
	deleteConsumerGroupViaCLI(t, env, id2)
	deleteConsumerGroupViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteConsumerGroupViaAdmin(t, id1)
		deleteConsumerGroupViaAdmin(t, id2)
		deleteConsumerGroupViaAdmin(t, id3)
	})

	basePlugin := `"plugins":{"limit-count":{"count":200,"time_window":60,"rejected_code":503,"key_type":"var","key":"remote_addr"}}`
	body1 := fmt.Sprintf(`{"id":"%s",%s,"labels":{"suite":"exp-cg"}}`, id1, basePlugin)
	body2 := fmt.Sprintf(`{"id":"%s",%s,"labels":{"suite":"exp-cg"}}`, id2, basePlugin)
	body3 := fmt.Sprintf(`{"id":"%s",%s,"labels":{"suite":"exp-cg-other"}}`, id3, basePlugin)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("consumer-group-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "consumer-group", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "consumer-group", "export", "--label", "suite=exp-cg", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "consumer-group-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "consumer-group", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_GlobalRule(t *testing.T) {
	id1 := "100001"
	id2 := "100002"
	id3 := "100003"
	env := setupGlobalRuleEnv(t)

	deleteGlobalRuleViaCLI(t, env, id1)
	deleteGlobalRuleViaCLI(t, env, id2)
	deleteGlobalRuleViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteGlobalRuleViaAdmin(t, id1)
		deleteGlobalRuleViaAdmin(t, id2)
		deleteGlobalRuleViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","plugins":{"prometheus":{}},"labels":{"suite":"exp-gr"}}`, id1)
	body2 := fmt.Sprintf(`{"id":"%s","plugins":{"prometheus":{}},"labels":{"suite":"exp-gr"}}`, id2)
	body3 := fmt.Sprintf(`{"id":"%s","plugins":{"prometheus":{}},"labels":{"suite":"exp-gr-other"}}`, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("global-rule-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "global-rule", "create", "-f", createFile)
		if err != nil && strings.Contains(stderr, "additional properties forbidden, found labels") {
			t.Skip("global-rule labels are not supported by current APISIX")
		}
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "global-rule", "export", "--label", "suite=exp-gr", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "global-rule", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "global-rule-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "global-rule", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_PluginConfig(t *testing.T) {
	id1 := "exp-pc-1"
	id2 := "exp-pc-2"
	id3 := "exp-pc-3"
	env := setupPluginConfigEnv(t)

	deletePluginConfigViaCLI(t, env, id1)
	deletePluginConfigViaCLI(t, env, id2)
	deletePluginConfigViaCLI(t, env, id3)
	t.Cleanup(func() {
		deletePluginConfigViaAdmin(t, id1)
		deletePluginConfigViaAdmin(t, id2)
		deletePluginConfigViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","plugins":{"prometheus":{}},"labels":{"suite":"exp-pc"}}`, id1)
	body2 := fmt.Sprintf(`{"id":"%s","plugins":{"prometheus":{}},"labels":{"suite":"exp-pc"}}`, id2)
	body3 := fmt.Sprintf(`{"id":"%s","plugins":{"prometheus":{}},"labels":{"suite":"exp-pc-other"}}`, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("plugin-config-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "plugin-config", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "plugin-config", "export", "--label", "suite=exp-pc", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "plugin-config-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "plugin-config", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_Proto(t *testing.T) {
	id1 := "exp-proto-1"
	id2 := "exp-proto-2"
	id3 := "exp-proto-3"
	env := setupProtoEnv(t)

	type protoPayload struct {
		ID      string            `json:"id"`
		Name    string            `json:"name"`
		Desc    string            `json:"desc"`
		Content string            `json:"content"`
		Labels  map[string]string `json:"labels"`
	}

	deleteProtoViaCLI(t, env, id1)
	deleteProtoViaCLI(t, env, id2)
	deleteProtoViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteProtoViaAdmin(t, id1)
		deleteProtoViaAdmin(t, id2)
		deleteProtoViaAdmin(t, id3)
	})

	protoContent := `syntax = "proto3";
package helloworld;
service Greeter {
    rpc SayHello (HelloRequest) returns (HelloReply) {}
}
message HelloRequest { string name = 1; }
message HelloReply { string message = 1; }`
	payloads := []protoPayload{
		{ID: id1, Name: id1, Desc: "proto export 1", Content: protoContent, Labels: map[string]string{"suite": "exp-proto"}},
		{ID: id2, Name: id2, Desc: "proto export 2", Content: protoContent, Labels: map[string]string{"suite": "exp-proto"}},
		{ID: id3, Name: id3, Desc: "proto export 3", Content: protoContent, Labels: map[string]string{"suite": "exp-proto-other"}},
	}

	for i, payload := range payloads {
		body, marshalErr := json.Marshal(payload)
		require.NoError(t, marshalErr)
		createFile := writeJSONFile(t, fmt.Sprintf("proto-export-%d.json", i+1), string(body))
		stdout, stderr, err := runA6WithEnv(env, "proto", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "proto", "export", "--label", "suite=exp-proto", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "proto", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "proto-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "proto", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_SSL(t *testing.T) {
	id1 := "exp-ssl-1"
	id2 := "exp-ssl-2"
	id3 := "exp-ssl-3"
	env := setupSSLEnv(t)

	deleteSSLViaCLI(t, env, id1)
	deleteSSLViaCLI(t, env, id2)
	deleteSSLViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteSSLViaAdmin(t, id1)
		deleteSSLViaAdmin(t, id2)
		deleteSSLViaAdmin(t, id3)
	})

	cert, key := readTestCert(t)
	certJSON := strings.ReplaceAll(cert, "\n", "\\n")
	keyJSON := strings.ReplaceAll(key, "\n", "\\n")

	body1 := fmt.Sprintf(`{"id":"%s","cert":"%s","key":"%s","snis":["exp-ssl-1.example.com"],"labels":{"suite":"exp-ssl"}}`, id1, certJSON, keyJSON)
	body2 := fmt.Sprintf(`{"id":"%s","cert":"%s","key":"%s","snis":["exp-ssl-2.example.com"],"labels":{"suite":"exp-ssl"}}`, id2, certJSON, keyJSON)
	body3 := fmt.Sprintf(`{"id":"%s","cert":"%s","key":"%s","snis":["exp-ssl-3.example.com"],"labels":{"suite":"exp-ssl-other"}}`, id3, certJSON, keyJSON)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("ssl-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "ssl", "export", "--label", "suite=exp-ssl", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "ssl", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "ssl-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "ssl", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestExport_StreamRoute(t *testing.T) {
	id1 := "exp-sr-1"
	id2 := "exp-sr-2"
	id3 := "exp-sr-3"
	env := setupStreamRouteEnv(t)

	deleteStreamRouteViaCLI(t, env, id1)
	deleteStreamRouteViaCLI(t, env, id2)
	deleteStreamRouteViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteStreamRouteViaAdmin(t, id1)
		deleteStreamRouteViaAdmin(t, id2)
		deleteStreamRouteViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","server_port":19101,"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"suite":"exp-sr"}}`, id1)
	body2 := fmt.Sprintf(`{"id":"%s","server_port":19102,"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"suite":"exp-sr"}}`, id2)
	body3 := fmt.Sprintf(`{"id":"%s","server_port":19103,"upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"suite":"exp-sr-other"}}`, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("stream-route-export-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "stream-route", "create", "-f", createFile)
		if err != nil && strings.Contains(stderr, "stream mode is disabled") {
			t.Skip("stream mode is disabled in current APISIX")
		}
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "stream-route", "export", "--label", "suite=exp-sr", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)

	stdout, stderr, err = runA6WithEnv(env, "stream-route", "export", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	allIDs := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, allIDs, id1)
	assert.Contains(t, allIDs, id2)
	assert.Contains(t, allIDs, id3)

	outFile := filepath.Join(t.TempDir(), "stream-route-export.yaml")
	stdout, stderr, err = runA6WithEnv(env, "stream-route", "export", "-f", outFile, "--output", "yaml")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	content, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), id1)
	assert.Contains(t, string(content), id2)
	assert.Contains(t, string(content), id3)
}

func TestLabelList_Route(t *testing.T) {
	id1 := "lbl-route-1"
	id2 := "lbl-route-2"
	id3 := "lbl-route-3"
	env := setupRouteEnv(t)

	deleteRouteViaCLI(t, env, id1)
	deleteRouteViaCLI(t, env, id2)
	deleteRouteViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteRouteViaAdmin(t, id1)
		deleteRouteViaAdmin(t, id2)
		deleteRouteViaAdmin(t, id3)
	})

	createLabeledRouteViaCLI(t, env, id1, "/lbl-route/1", "env", "test")
	createLabeledRouteViaCLI(t, env, id2, "/lbl-route/2", "env", "test")
	createLabeledRouteViaCLI(t, env, id3, "/lbl-route/3", "env", "prod")

	stdout, stderr, err := runA6WithEnv(env, "route", "list", "--label", "env=test", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)
}

func TestLabelList_Upstream(t *testing.T) {
	id1 := "lbl-up-1"
	id2 := "lbl-up-2"
	id3 := "lbl-up-3"
	env := setupUpstreamEnv(t)

	deleteUpstreamViaCLI(t, env, id1)
	deleteUpstreamViaCLI(t, env, id2)
	deleteUpstreamViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteUpstreamViaAdmin(t, id1)
		deleteUpstreamViaAdmin(t, id2)
		deleteUpstreamViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"test"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"test"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"prod"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("upstream-label-list-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "upstream", "list", "--label", "env=test", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)
}

func TestLabelList_Service(t *testing.T) {
	id1 := "lbl-svc-1"
	id2 := "lbl-svc-2"
	id3 := "lbl-svc-3"
	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, id1)
	deleteServiceViaCLI(t, env, id2)
	deleteServiceViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteServiceViaAdmin(t, id1)
		deleteServiceViaAdmin(t, id2)
		deleteServiceViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"test"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"test"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"prod"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("service-label-list-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "service", "list", "--label", "env=test", "--output", "json")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	ids := strings.Join(collectFieldValues(parseJSONArrayOutput(t, stdout), "id"), ",")
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.NotContains(t, ids, id3)
}

func TestLabelDelete_Upstream(t *testing.T) {
	id1 := "ldel-up-1"
	id2 := "ldel-up-2"
	id3 := "ldel-up-3"
	env := setupUpstreamEnv(t)

	deleteUpstreamViaCLI(t, env, id1)
	deleteUpstreamViaCLI(t, env, id2)
	deleteUpstreamViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteUpstreamViaAdmin(t, id1)
		deleteUpstreamViaAdmin(t, id2)
		deleteUpstreamViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"test"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"test"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"id":"%s","name":"%s","type":"roundrobin","nodes":{"127.0.0.1:8080":1},"labels":{"env":"prod"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("upstream-label-delete-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "upstream", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "upstream", "delete", "--label", "env=test", "--force")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "upstream", "get", id1)
	assert.Error(t, err)
	_, stderr, err = runA6WithEnv(env, "upstream", "get", id2)
	assert.Error(t, err)
	stdout, stderr, err = runA6WithEnv(env, "upstream", "get", id3)
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, id3)
}

func TestLabelDelete_Service(t *testing.T) {
	id1 := "ldel-svc-1"
	id2 := "ldel-svc-2"
	id3 := "ldel-svc-3"
	env := setupServiceEnv(t)

	deleteServiceViaCLI(t, env, id1)
	deleteServiceViaCLI(t, env, id2)
	deleteServiceViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteServiceViaAdmin(t, id1)
		deleteServiceViaAdmin(t, id2)
		deleteServiceViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"test"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"test"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"id":"%s","name":"%s","upstream":{"type":"roundrobin","nodes":{"127.0.0.1:8080":1}},"labels":{"env":"prod"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("service-label-delete-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "service", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "service", "delete", "--label", "env=test", "--force")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "service", "get", id1)
	assert.Error(t, err)
	_, stderr, err = runA6WithEnv(env, "service", "get", id2)
	assert.Error(t, err)
	stdout, stderr, err = runA6WithEnv(env, "service", "get", id3)
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, id3)
}

func TestLabelDelete_Consumer(t *testing.T) {
	id1 := "ldel-consumer-1"
	id2 := "ldel-consumer-2"
	id3 := "ldel-consumer-3"
	env := setupConsumerEnv(t)

	deleteConsumerViaCLI(t, env, id1)
	deleteConsumerViaCLI(t, env, id2)
	deleteConsumerViaCLI(t, env, id3)
	t.Cleanup(func() {
		deleteConsumerViaAdmin(t, id1)
		deleteConsumerViaAdmin(t, id2)
		deleteConsumerViaAdmin(t, id3)
	})

	body1 := fmt.Sprintf(`{"username":"%s","plugins":{"key-auth":{"key":"%s-key"}},"labels":{"env":"test"}}`, id1, id1)
	body2 := fmt.Sprintf(`{"username":"%s","plugins":{"key-auth":{"key":"%s-key"}},"labels":{"env":"test"}}`, id2, id2)
	body3 := fmt.Sprintf(`{"username":"%s","plugins":{"key-auth":{"key":"%s-key"}},"labels":{"env":"prod"}}`, id3, id3)

	for i, body := range []string{body1, body2, body3} {
		createFile := writeJSONFile(t, fmt.Sprintf("consumer-label-delete-%d.json", i+1), body)
		stdout, stderr, err := runA6WithEnv(env, "consumer", "create", "-f", createFile)
		require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	}

	stdout, stderr, err := runA6WithEnv(env, "consumer", "delete", "--label", "env=test", "--force")
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)

	_, stderr, err = runA6WithEnv(env, "consumer", "get", id1)
	assert.Error(t, err)
	_, stderr, err = runA6WithEnv(env, "consumer", "get", id2)
	assert.Error(t, err)
	stdout, stderr, err = runA6WithEnv(env, "consumer", "get", id3)
	require.NoError(t, err, "stdout=%s stderr=%s", stdout, stderr)
	assert.Contains(t, stdout, id3)
}
