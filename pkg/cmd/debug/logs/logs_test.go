package logs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogs_DetectContainerParsing(t *testing.T) {
	parsed := parseDockerPSNames("apisix\napisix-dashboard\n\n")
	require.Len(t, parsed, 2)
	assert.Equal(t, "apisix", parsed[0])
	assert.Equal(t, "apisix-dashboard", parsed[1])
}

func TestLogs_BuildDockerArgs(t *testing.T) {
	opts := &Options{Tail: 100}
	args := buildDockerArgs(opts, "apisix")
	assert.Equal(t, []string{"logs", "--tail", "100", "apisix"}, args)
}

func TestLogs_BuildDockerArgsWithFollow(t *testing.T) {
	opts := &Options{Follow: true, Tail: 10}
	args := buildDockerArgs(opts, "apisix")
	assert.Equal(t, []string{"logs", "--follow", "--tail", "10", "apisix"}, args)
}

func TestLogs_BuildDockerArgsWithSince(t *testing.T) {
	opts := &Options{Tail: 5, Since: "1h"}
	args := buildDockerArgs(opts, "apisix")
	assert.Equal(t, []string{"logs", "--tail", "5", "--since", "1h", "apisix"}, args)
}

func TestLogs_ReadLastLines(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "apisix.log")
	content := "line-1\nline-2\nline-3\nline-4\n"
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	lines, err := readLastLines(filePath, 2)
	require.NoError(t, err)
	assert.Equal(t, []string{"line-3", "line-4"}, lines)
}

func TestLogs_NoContainerNoFile(t *testing.T) {
	_, err := chooseContainer(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no APISIX container found")
}

func TestLogs_ChooseContainerSingle(t *testing.T) {
	container, err := chooseContainer([]string{"apisix"})
	require.NoError(t, err)
	assert.Equal(t, "apisix", container)
}

func TestLogs_ChooseContainerMultiple(t *testing.T) {
	_, err := chooseContainer([]string{"apisix", "apisix-staging"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple APISIX containers found")
	assert.Contains(t, err.Error(), "apisix, apisix-staging")
}
