package update

import (
	"testing"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginMetadataUpdate_MissingFile(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &cmd.Factory{IOStreams: ios}
	c := NewCmdUpdate(f)
	c.SetArgs([]string{"syslog"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "file")
}
