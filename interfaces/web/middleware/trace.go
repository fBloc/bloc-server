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

		spanID := value_object.NewUUID().String()
		if traceID == "" {
			traceID = spanID
		}

		ctx := context.WithValue(r.Context(), web.RequestContextTraceID, traceID)
		ctx = context.WithValue(ctx, web.RequestContextSpanID, spanID)
		ctx = context.WithValue(ctx, web.RequestContextParentSpanID, parentSpanID)
		r = r.WithContext(ctx)

		h(w, r, ps)
	}
}
