package value_object

import (
	"context"

	"github.com/spf13/cast"
)

type TraceFlag string

const (
	TraceID      TraceFlag = "trace_id"
	SpanID       TraceFlag = "span_id"
	ParentSpanID TraceFlag = "parent_span_id"
)

func NewTraceID() string {
	return NewUUID().String()
}

func NewSpanID() string {
	return NewUUID().String()
}

func NewTraceTags() map[string]string {
	spanID := NewSpanID()
	return map[string]string{
		string(TraceID):      spanID,
		string(SpanID):       spanID,
		string(ParentSpanID): ""}
}

func SetTraceIDToContext(traceID string) context.Context {
	return context.WithValue(context.Background(), TraceID, traceID)
}

func GetTraceIDFromContext(ctx context.Context) string {
	val := ctx.Value(TraceID)
	if val == nil {
		return ""
	}
	return cast.ToString(val)
}
