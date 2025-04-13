package util

import (
	"encoding/json"

	"github/M2A96/Monopoly.git/object"

	"go.opentelemetry.io/otel/trace"
)

type spanContext struct {
	trace.SpanContext
}

var (
	_ json.Marshaler   = (*spanContext)(nil)
	_ object.GetMapper = (*spanContext)(nil)
)

// NewSpanContext is a function.
func NewSpanContext(
	traceSpan trace.Span,
) *spanContext {
	if traceSpan == nil {
		return &spanContext{
			SpanContext: trace.SpanContext{},
		}
	}

	return &spanContext{
		SpanContext: traceSpan.SpanContext(),
	}
}

// GetMap is a function.
func (util *spanContext) GetMap() map[string]any {
	return map[string]any{
		"trace_id":    util.TraceID(),
		"span_id":     util.SpanID(),
		"trace_flags": util.TraceFlags(),
		"trace_state": util.TraceState(),
		"remote":      util.IsRemote(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (util *spanContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(util.GetMap())
}
