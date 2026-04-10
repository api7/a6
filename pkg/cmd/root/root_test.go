package root

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/api7/a6/internal/extension"
	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
)

func TestNewCmdRoot_RegistersCoreCommandsAndFlags(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	rootCmd := NewCmdRoot(&cmd.Factory{IOStreams: ios})

	assert.Equal(t, "a6", rootCmd.Use)
	assert.Equal(t, "Apache APISIX CLI", rootCmd.Short)
	assert.True(t, rootCmd.SilenceUsage)
	assert.True(t, rootCmd.SilenceErrors)

	require.NotNil(t, rootCmd.PersistentFlags().Lookup("output"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("context"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("server"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("api-key"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("force"))

	for _, name := range []string{
		"config",
		"consumer",
		"debug",
		"extension",
		"global-rule",
		"plugin-config",
		"route",
		"service",
		"upstream",
		"version",
	} {
		found, _, err := rootCmd.Find([]string{name})
		require.NoError(t, err)
		assert.NotNil(t, found)
	}
}

func TestNewCmdRoot_LoadsAndExecutesExtensionCommands(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-based extension execution test is unix-specific")
	}

	configDir := t.TempDir()
	t.Setenv("A6_CONFIG_DIR", configDir)

	resultPath := filepath.Join(t.TempDir(), "extension-result.txt")
	t.Setenv("A6_ROOT_TEST_OUTPUT", resultPath)

	extDir := filepath.Join(configDir, "extensions", "a6-hello")
	binaryName := "a6-hello"
	binaryPath := filepath.Join(extDir, binaryName)
	require.NoError(t, os.MkdirAll(extDir, 0o755))
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\nprintf '%s' \"$*\" > \"$A6_ROOT_TEST_OUTPUT\"\n"), 0o755))
	require.NoError(t, writeManifest(filepath.Join(extDir, "manifest.yaml"), extension.Manifest{
		Name:        "hello",
		Owner:       "api7",
		Repo:        "a6-hello",
		Version:     "1.0.0",
		Description: "test extension",
		BinaryPath:  binaryName,
	}))

	ios, _, _, _ := iostreams.Test()
	rootCmd := NewCmdRoot(&cmd.Factory{IOStreams: ios})
	rootCmd.SetArgs([]string{"hello", "arg-one", "arg-two"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(resultPath)
	require.NoError(t, err)
	assert.Equal(t, "arg-one arg-two", strings.TrimSpace(string(content)))

	cmd, _, err := rootCmd.Find([]string{"hello"})
	require.NoError(t, err)
	assert.Equal(t, "extension", cmd.GroupID)
	assert.Equal(t, "test extension", cmd.Short)
}

func writeManifest(path string, manifest extension.Manifest) error {
	content, err := yaml.Marshal(&manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
