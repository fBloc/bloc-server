package web

type RequestContextKey = string

const (
	RequestContextUserKey      RequestContextKey = "req_user"
	RequestContextTraceID      RequestContextKey = "trace_id"
	RequestContextSpanID       RequestContextKey = "span_id"
	RequestContextParentSpanID RequestContextKey = "parent_span_id"
)
