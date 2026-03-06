package cmdutil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExporter_JSON(t *testing.T) {
	buf := &bytes.Buffer{}
	exp := NewExporter("json", buf)

	data := map[string]string{"name": "test-route", "uri": "/api"}
	err := exp.Write(data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"name": "test-route"`)
	assert.Contains(t, output, `"uri": "/api"`)
}

func TestExporter_YAML(t *testing.T) {
	buf := &bytes.Buffer{}
	exp := NewExporter("yaml", buf)

	data := map[string]string{"name": "test-route", "uri": "/api"}
	err := exp.Write(data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "name: test-route")
	assert.Contains(t, output, "uri: /api")
}

func TestExporter_JSONArray(t *testing.T) {
	buf := &bytes.Buffer{}
	exp := NewExporter("json", buf)

	data := []map[string]string{
		{"id": "1", "name": "route-a"},
		{"id": "2", "name": "route-b"},
	}
	err := exp.Write(data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"id": "1"`)
	assert.Contains(t, output, `"id": "2"`)
}

func TestExporter_UnsupportedFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	exp := NewExporter("xml", buf)

	err := exp.Write(map[string]string{"key": "val"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format: xml")
}

func TestExporter_JSONPrettyPrinted(t *testing.T) {
	buf := &bytes.Buffer{}
	exp := NewExporter("json", buf)

	data := map[string]string{"key": "value"}
	err := exp.Write(data)
	require.NoError(t, err)

	// Verify pretty printing (indented)
	assert.Contains(t, buf.String(), "  \"key\"")
}
