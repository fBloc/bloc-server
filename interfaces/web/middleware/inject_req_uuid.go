package middleware

import (
	"context"
	"net/http"

	"github.com/fBloc/bloc-backend-go/interfaces/web"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// InjectReqUUID 插入用于串联日志的uuid
func InjectReqUUID(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), web.RequestContextUUIDKey, uuid.New().String())
		r = r.WithContext(ctx)
		h(w, r, ps)
	}
}
