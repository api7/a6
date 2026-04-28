package selector

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var stdinSwapMu sync.Mutex

func TestSelectOne_EmptyItems(t *testing.T) {
	_, err := SelectOne("Select a route", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no items available")
}

func TestSelectOne_AllItemsMissingIDs(t *testing.T) {
	items := []Item{
		{ID: "", Label: "first"},
		{ID: "", Label: "second"},
	}

	_, err := SelectOne("Select a route", items)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no items available")
}

func TestSelectOne_RequiresTerminal(t *testing.T) {
	stdinSwapMu.Lock()
	t.Cleanup(stdinSwapMu.Unlock)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "stdin.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte("not-a-tty"), 0o644))

	f, err := os.Open(inputPath)
	require.NoError(t, err)
	defer f.Close()

	original := os.Stdin
	t.Cleanup(func() {
		os.Stdin = original
	})
	os.Stdin = f

	_, err = SelectOne("Select a route", []Item{{ID: "route-1", Label: "Route 1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "interactive selection requires a terminal")
}
