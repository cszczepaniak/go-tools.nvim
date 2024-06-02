package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"
)

type textHandler struct {
	w     io.Writer
	attrs []slog.Attr
}

func (t textHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (t textHandler) Handle(_ context.Context, r slog.Record) error {
	fmt.Fprintf(t.w, "[%s]", r.Time.Format(time.DateTime))
	fmt.Fprintf(t.w, " %s", r.Level)

	r.AddAttrs(t.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(t.w, " %s", a)
		return true
	})

	fmt.Fprintf(t.w, " %s\n", r.Message)
	return nil
}

func (t textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	t.attrs = append(t.attrs, attrs...)
	return t
}

func (t textHandler) WithGroup(name string) slog.Handler {
	return t
}
