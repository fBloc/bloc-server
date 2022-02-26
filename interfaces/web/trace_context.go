package web

import (
	"context"
)

func GetReqTraceIDFromContext(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(RequestContextTraceID).(string)
	return traceID, ok
}

func GetReqSpanIDFromContext(ctx context.Context) (string, bool) {
	spanID, ok := ctx.Value(RequestContextSpanID).(string)
	return spanID, ok
}

func GetReqParentSpanIDFromContext(ctx context.Context) (string, bool) {
	parentSpanID, ok := ctx.Value(RequestContextParentSpanID).(string)
	return parentSpanID, ok
}

func GetTraceAboutFields(ctx context.Context) map[string]string {
	return map[string]string{
		RequestContextTraceID:      ctx.Value(RequestContextTraceID).(string),
		RequestContextSpanID:       ctx.Value(RequestContextSpanID).(string),
		RequestContextParentSpanID: ctx.Value(RequestContextParentSpanID).(string),
	}
}
