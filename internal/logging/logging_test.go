package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	e := WithFields(map[string]any{"a": 123})
	assert.Equal(t, map[string]any{"a": 123}, e.fields)
	assert.Equal(t, []any{"a", 123}, e.getKVs())
}
