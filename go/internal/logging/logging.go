package logging

import (
	"io"

	"github.com/charmbracelet/log"
)

var logger *log.Logger

func InitLogger(w io.Writer) {
	logger = log.New(w)
}

type loggingEntry struct {
	l      *log.Logger
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

func (l loggingEntry) getKVs() []any {
	kvs := make([]any, 0, 2*len(l.fields))
	for k, v := range kvs {
		kvs = append(kvs, k, v)
	}
	return kvs
}
