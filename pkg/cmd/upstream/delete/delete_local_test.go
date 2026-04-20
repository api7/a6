package delete

import (
	"testing"

	"github.com/api7/a6/pkg/cmd"
	"github.com/api7/a6/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamDelete_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := deleteRun(&Options{IO: ios})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}

func TestUpstreamDelete_AllAndLabelMutuallyExclusive(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: ios}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"--all", "--label", "env=test"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--all and --label are mutually exclusive")
}

func TestUpstreamDelete_AllWithID(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: ios}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"1", "--all"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--all cannot be used with a specific ID")
}

func TestUpstreamDelete_LabelWithID(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmd.Factory{IOStreams: ios}

	c := NewCmdDelete(f)
	c.SetArgs([]string{"1", "--label", "env=test"})
	err := c.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--label cannot be used with a specific ID")
}
