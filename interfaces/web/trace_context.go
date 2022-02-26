package web

import (
	"context"

	"github.com/fBloc/bloc-server/value_object"
)

func GetReqTraceIDFromContext(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(value_object.TraceID).(string)
	return traceID, ok
}

func GetReqSpanIDFromContext(ctx context.Context) (string, bool) {
	spanID, ok := ctx.Value(value_object.SpanID).(string)
	return spanID, ok
}

func GetReqParentSpanIDFromContext(ctx context.Context) (string, bool) {
	parentSpanID, ok := ctx.Value(value_object.ParentSpanID).(string)
	return parentSpanID, ok
}

func GetTraceAboutFields(ctx context.Context) map[string]string {
	return map[string]string{
		string(value_object.TraceID):      ctx.Value(value_object.TraceID).(string),
		string(value_object.SpanID):       ctx.Value(value_object.SpanID).(string),
		string(value_object.ParentSpanID): ctx.Value(value_object.ParentSpanID).(string),
	}
}
