package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
)

// NewDefaultHandler creates a new default handler that writes human-readable log records to the given output.
func NewDefaultHandler(output io.Writer) slog.Handler {
	return &defaultHandler{output: output}
}

// defaultHandler is a slog.Handler that writes human-readable log records to an output.
type defaultHandler struct {
	output io.Writer
}

// Enabled returns true for all levels as we handle the cutoff ourselves using reactive variables and the ability to
// set loggers to nil.
func (h *defaultHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle writes the log record to the output.
func (h *defaultHandler) Handle(_ context.Context, r slog.Record) error {
	var (
		namespace    string
		fieldsBuffer = new(bytes.Buffer)
	)

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key == namespaceKey {
			namespace = attr.Value.Any().(string)
		} else {
			fieldsBuffer.WriteByte(' ')
			fieldsBuffer.WriteString(attr.String())
		}

		return true
	})

	fmt.Fprintf(h.output, "%s\t%s\t%s\t%s\t%s\n", r.Time.Format("2006/01/02 15:04:05"), LevelName(r.Level), namespace, r.Message, fieldsBuffer.String())

	return nil
}

// WithAttrs is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (h *defaultHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	panic("not supported")
}

// WithGroup is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (h *defaultHandler) WithGroup(_ string) slog.Handler {
	return h
}
