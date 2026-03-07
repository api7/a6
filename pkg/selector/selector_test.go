package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectOne_EmptyItems(t *testing.T) {
	_, err := SelectOne("Select a route", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no items available")
}
