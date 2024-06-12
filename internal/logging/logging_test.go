package logging

import (
	"testing"

	"github.com/shoenig/test"
)

func TestLogging(t *testing.T) {
	e := WithFields(map[string]any{"a": 123})
	test.Eq(t, map[string]any{"a": 123}, e.fields)
	test.Eq(t, []any{"a", 123}, e.getKVs())
}
