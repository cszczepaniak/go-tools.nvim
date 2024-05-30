package linewriter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineWriter_Write(t *testing.T) {
	lw := &Writer{}

	fmt.Fprint(lw, "abc")
	fmt.Fprint(lw, "\n")
	fmt.Fprintf(lw, "%s\n\n%d", "def", 123)

	lns := lw.TakeLines()
	assert.Equal(t, []string{
		"abc",
		"def",
		"",
		"123",
	}, lns)

	lw = &Writer{}

	fmt.Fprint(lw, "abc\n")

	lns = lw.TakeLines()
	assert.Equal(t, []string{
		"abc",
		"",
	}, lns)
}
