package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"env=test", "env:test"},
		{"env=prod", "env:prod"},
		{"team=backend", "team:backend"},
		{"env:test", "env:test"},
		{"env", "env"},
		{"key=value=extra", "key:value=extra"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, NormalizeLabel(tc.input))
		})
	}
}
