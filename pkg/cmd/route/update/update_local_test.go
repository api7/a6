package update

import (
	"testing"

	"github.com/api7/a6/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteUpdate_NoArgsNonTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := updateRun(&Options{IO: ios})
	require.Error(t, err)
	assert.Equal(t, "id argument is required (or run interactively in a terminal)", err.Error())
}
