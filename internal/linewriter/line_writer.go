package linewriter

import (
	"bytes"
	"fmt"
)

type Writer struct {
	hasLeftover bool
	curr        []byte
	lns         []string
}

func (lw *Writer) WriteLinef(f string, args ...any) *Writer {
	lw.lns = append(lw.lns, fmt.Sprintf(f, args...))
	return lw
}

func (lw *Writer) Write(bs []byte) (int, error) {
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

func (lw *Writer) Flush() {
	if lw.hasLeftover {
		lw.lns = append(lw.lns, string(lw.curr))
		lw.hasLeftover = false
		lw.curr = lw.curr[:0]
	}
}

func (lw *Writer) TakeLines() []string {
	lw.Flush()
	return lw.lns
}
