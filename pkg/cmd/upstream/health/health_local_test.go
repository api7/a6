package health

import (
	"testing"

	"github.com/api7/a6/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamHealth_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := healthRun(&Options{IO: ios})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}

func TestDeriveControlURL(t *testing.T) {
	controlURL, err := deriveControlURL("http://127.0.0.1:9180")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:9090", controlURL)
}

func TestBuildHealthURL(t *testing.T) {
	healthURL, err := buildHealthURL("http://127.0.0.1:19090", "u-1")
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:19090/v1/healthcheck/upstreams/u-1", healthURL)
}

func TestBuildHealthURL_MissingHost(t *testing.T) {
	_, err := buildHealthURL("http://", "u-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing host")
}
