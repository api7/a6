package tableprinter

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestRender_EmptyTable(t *testing.T) {
	var out bytes.Buffer
	p := New(&out)

	require.NoError(t, p.Render())
	assert.Empty(t, out.String())
	assert.Equal(t, 0, p.RowCount())
}

func TestRender_WithHeadersAndRows(t *testing.T) {
	var out bytes.Buffer
	p := New(&out)
	p.SetHeaders("ID", "NAME")
	p.AddRow("1", "route-a")
	p.AddRow("2", "route-b")

	require.NoError(t, p.Render())

	rendered := out.String()
	assert.Contains(t, rendered, "ID")
	assert.Contains(t, rendered, "NAME")
	assert.Contains(t, rendered, "route-a")
	assert.Contains(t, rendered, "route-b")
	assert.Equal(t, 2, p.RowCount())
}

func TestRender_ReturnsFlushError(t *testing.T) {
	p := New(failingWriter{})
	p.SetHeaders("ID")

	err := p.Render()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}
