package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve_GoInstallModuleVersion(t *testing.T) {
	// `go install .../cmd/a6@v1.0.0` embeds the module version but no ldflags.
	info := &debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}}

	v, c, d := resolve("dev", "unknown", "unknown", info)

	assert.Equal(t, "v1.0.0", v)
	assert.Equal(t, "unknown", c)
	assert.Equal(t, "unknown", d)
}

func TestResolve_DevelBuildStaysDev(t *testing.T) {
	// `go build` from a source checkout reports "(devel)" as the module version.
	info := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}

	v, _, _ := resolve("dev", "unknown", "unknown", info)

	assert.Equal(t, "dev", v)
}

func TestResolve_LdflagsTakePrecedence(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v9.9.9"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "ffffffffffffffffffffffffffffffffffffffff"},
			{Key: "vcs.time", Value: "2026-01-01T00:00:00Z"},
		},
	}

	v, c, d := resolve("v1.2.3", "abc1234", "2026-05-27T03:23:51Z", info)

	assert.Equal(t, "v1.2.3", v)
	assert.Equal(t, "abc1234", c)
	assert.Equal(t, "2026-05-27T03:23:51Z", d)
}

func TestResolve_VCSSettingsFallback(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef0123456789abcdef01234567"},
			{Key: "vcs.time", Value: "2026-06-03T00:00:00Z"},
		},
	}

	v, c, d := resolve("dev", "unknown", "unknown", info)

	assert.Equal(t, "dev", v)
	assert.Equal(t, "0123456", c, "commit should be shortened to 7 characters")
	assert.Equal(t, "2026-06-03T00:00:00Z", d)
}

func TestResolve_EmptyBuildInfo(t *testing.T) {
	v, c, d := resolve("dev", "unknown", "unknown", &debug.BuildInfo{})

	assert.Equal(t, "dev", v)
	assert.Equal(t, "unknown", c)
	assert.Equal(t, "unknown", d)
}
