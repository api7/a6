package iostreams

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTest_ReturnsBuffers(t *testing.T) {
	ios, in, out, errBuf := Test()
	require.NotNil(t, ios)
	require.NotNil(t, in)
	require.NotNil(t, out)
	require.NotNil(t, errBuf)
}

func TestTest_DefaultNotTTY(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.IsStdinTTY())
	assert.False(t, ios.IsStdoutTTY())
	assert.False(t, ios.IsStderrTTY())
}

func TestSetStdoutTTY(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.IsStdoutTTY())

	ios.SetStdoutTTY(true)
	assert.True(t, ios.IsStdoutTTY())

	ios.SetStdoutTTY(false)
	assert.False(t, ios.IsStdoutTTY())
}

func TestSetStdinTTY(t *testing.T) {
	ios, _, _, _ := Test()
	ios.SetStdinTTY(true)
	assert.True(t, ios.IsStdinTTY())
}

func TestSetStderrTTY(t *testing.T) {
	ios, _, _, _ := Test()
	ios.SetStderrTTY(true)
	assert.True(t, ios.IsStderrTTY())
}

func TestColorEnabled_DisabledByDefault(t *testing.T) {
	ios, _, _, _ := Test()
	// Test IOStreams has outTTY=false, so color should be disabled
	assert.False(t, ios.ColorEnabled())
}

func TestColorEnabled_EnabledWhenTTY(t *testing.T) {
	ios, _, _, _ := Test()
	ios.SetStdoutTTY(true)

	t.Setenv("NO_COLOR", "")
	assert.True(t, ios.ColorEnabled())
}

func TestColorEnabled_DisabledByEnv(t *testing.T) {
	ios, _, _, _ := Test()
	ios.SetStdoutTTY(true)

	t.Setenv("NO_COLOR", "1")
	assert.False(t, ios.ColorEnabled())
}

func TestTest_WriteAndRead(t *testing.T) {
	ios, _, out, errBuf := Test()

	_, err := ios.Out.Write([]byte("hello stdout"))
	require.NoError(t, err)
	assert.Equal(t, "hello stdout", out.String())

	_, err = ios.ErrOut.Write([]byte("hello stderr"))
	require.NoError(t, err)
	assert.Equal(t, "hello stderr", errBuf.String())
}

func TestSystem_NotNil(t *testing.T) {
	ios := System()
	assert.NotNil(t, ios.In)
	assert.NotNil(t, ios.Out)
	assert.NotNil(t, ios.ErrOut)
}
