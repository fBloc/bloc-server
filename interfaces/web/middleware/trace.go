package middleware

import (
	"context"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

// WithTrace trace logs by trace_id & span_id & parent_span_id
func WithTrace(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		traceID := r.Header.Get(web.RequestContextTraceID)
		parentSpanID := r.Header.Get(web.RequestContextSpanID)

		spanID := value_object.NewSpanID()
		if traceID == "" {
			traceID = spanID
		}

		ctx := context.WithValue(r.Context(), value_object.TraceID, traceID)
		ctx = context.WithValue(ctx, value_object.SpanID, spanID)
		ctx = context.WithValue(ctx, value_object.ParentSpanID, parentSpanID)
		r = r.WithContext(ctx)

		h(w, r, ps)
	}
}
