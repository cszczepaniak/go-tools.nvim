package logging

import (
	"io"
	"log/slog"
	"os"
)

var logger *slog.Logger

func InitLogger(w io.Writer) {
	logger = slog.New(
		slog.NewTextHandler(w, &slog.HandlerOptions{}),
	)
}

func WithError(err error) loggingEntry {
	e := loggingEntry{fields: make(map[string]any)}
	return e.WithError(err)
}

func WithFields(fields map[string]any) loggingEntry {
	return loggingEntry{fields: fields}
}

func Debug(msg string) {
	logger.Debug(msg)
}

func Info(msg string) {
	logger.Info(msg)
}

func Warn(msg string) {
	logger.Warn(msg)
}

func Error(msg string) {
	logger.Error(msg)
}

// Fatal logs at the error level and then exits with code 1.
func Fatal(msg string) {
	logger.Error(msg)
	os.Exit(1)
}

type loggingEntry struct {
	fields map[string]any
}

func (l loggingEntry) WithError(err error) loggingEntry {
	l.fields["error"] = err
	return l
}

func (l loggingEntry) WithFields(fields map[string]any) loggingEntry {
	for k, v := range fields {
		l.fields[k] = v
	}
	return l
}

func (l loggingEntry) Debug(msg string) {
	logger.Debug(msg, l.getKVs()...)
}

func (l loggingEntry) Info(msg string) {
	logger.Info(msg, l.getKVs()...)
}

func (l loggingEntry) Warn(msg string) {
	logger.Warn(msg, l.getKVs()...)
}

func (l loggingEntry) Error(msg string) {
	logger.Error(msg, l.getKVs()...)
}

// Fatal logs at the error level and then exits with code 1.
func (l loggingEntry) Fatal(msg string) {
	logger.Error(msg, l.getKVs()...)
	os.Exit(1)
}

func (l loggingEntry) getKVs() []any {
	kvs := make([]any, 0, 2*len(l.fields))
	for k, v := range kvs {
		kvs = append(kvs, k, v)
	}
	return kvs
}
