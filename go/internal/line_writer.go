package internal

import (
	"bytes"
	"fmt"
)

type LineWriter struct {
	hasLeftover bool
	curr        []byte
	lns         []string
}

func (lw *LineWriter) WriteLinef(f string, args ...any) *LineWriter {
	lw.lns = append(lw.lns, fmt.Sprintf(f, args...))
	return lw
}

func (lw *LineWriter) Write(bs []byte) (int, error) {
	lw.hasLeftover = false
	for {
		idx := bytes.IndexByte(bs, '\n')
		if idx == -1 {
			lw.hasLeftover = true
			lw.curr = append(lw.curr, bs...)
			break
		}

		lw.curr = append(lw.curr, bs[:idx]...)
		lw.lns = append(lw.lns, string(lw.curr))
		lw.curr = lw.curr[:0]
		bs = bs[idx+1:]
	}

	return len(bs), nil
}

func (lw *LineWriter) Flush() {
	if lw.hasLeftover {
		lw.lns = append(lw.lns, string(lw.curr))
		lw.hasLeftover = false
		lw.curr = lw.curr[:0]
	}
}

func (lw *LineWriter) TakeLines() []string {
	lw.Flush()
	return lw.lns
}
