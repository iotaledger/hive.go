package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// NewTextHandler creates a new handler that writes human-readable log records to the given output.
func NewTextHandler(options *Options) slog.Handler {
	t := &textHandler{
		output:     options.Output,
		timeFormat: options.TimeFormat,
		ioWorker:   workerpool.New("log.TextHandler", workerpool.WithWorkerCount(1)).Start(),
	}

	formatString := "%s\t%-7s\t%s\t%s %s\n"
	t.formatString.Store(&formatString)

	return t
}

// textHandler is a slog.Handler that writes human-readable log records to an output.
type textHandler struct {
	output             io.Writer
	timeFormat         string
	maxNamespaceLength atomic.Int64
	formatString       atomic.Pointer[string]
	updateMutex        sync.Mutex
	ioWorker           *workerpool.WorkerPool
}

// Enabled returns true for all levels as we handle the cutoff ourselves using reactive variables and the ability to
// set loggers to nil.
func (t *textHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle writes the log record to the output.
func (t *textHandler) Handle(_ context.Context, r slog.Record) error {
	t.ioWorker.Submit(func() {
		var namespace string
		fieldsBuffer := new(bytes.Buffer)

		fieldCount := r.NumAttrs() - 1
		if fieldCount > 0 {
			fieldsBuffer.WriteString("(")
		}

		r.Attrs(func(attr slog.Attr) bool {
			if attr.Key == namespaceKey {
				//nolint:forcetypeassert // false positive, we know that the value is a string for a namespace key
				namespace = attr.Value.Any().(string)
			} else {
				fieldsBuffer.WriteString(attr.String())
				fieldsBuffer.WriteString(" ")
			}

			return true
		})

		if fieldCount > 0 {
			fieldsBuffer.Truncate(fieldsBuffer.Len() - 1)
			fieldsBuffer.WriteString(")")
		}

		if _, err := fmt.Fprintf(t.output, t.buildFormatString(namespace), r.Time.Format(t.timeFormat), LevelName(r.Level), namespace, r.Message, fieldsBuffer.String()); err != nil {
			panic(ierrors.Wrap(err, "writing log record failed"))
		}
	})

	return nil
}

// WithAttrs is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (t *textHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	panic("not supported")
}

// WithGroup is not supported (we don't want to support contextual logging where we pass around loggers between code
// parts but rather have a strictly hierarchical logging based on derived namespaces).
func (t *textHandler) WithGroup(_ string) slog.Handler {
	panic("not supported")
}

func (t *textHandler) buildFormatString(namespace string) string {
	currentMaxNamespaceLength := int(t.maxNamespaceLength.Load())
	currentFormatString := *(t.formatString.Load())

	if namespaceLength := len(namespace); namespaceLength > currentMaxNamespaceLength {
		t.updateMutex.Lock()
		defer t.updateMutex.Unlock()

		if namespaceLength > int(t.maxNamespaceLength.Load()) {
			currentFormatString = "%s\t%-7s\t%-" + strconv.Itoa(namespaceLength) + "s\t%s %s\n"

			t.formatString.Store(&currentFormatString)
			t.maxNamespaceLength.Store(int64(namespaceLength))
		}
	}

	return currentFormatString
}
