package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RespMsg struct {
	Code    int         `json:"status_code"`
	Msg     string      `json:"status_msg"`
	Data    interface{} `json:"data"`
	TraceID string      `json:"trace_id"`
}

func WriteSucResp(w *http.ResponseWriter, r *http.Request, data interface{}) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())
	resp := RespMsg{
		Code:    http.StatusOK,
		Data:    data,
		TraceID: traceID}
	(*w).Write(resp.JSONBytes())
}

func WriteDeleteSucResp(w *http.ResponseWriter, r *http.Request, delAmount int64) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())
	resp := RespMsg{
		Code:    http.StatusOK,
		Data:    map[string]int64{"delete_amount": delAmount},
		TraceID: traceID,
	}
	(*w).Write(resp.JSONBytes())
}

func WritePlainSucOkResp(w *http.ResponseWriter, r *http.Request) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())
	resp := RespMsg{Code: http.StatusOK, Data: "ok", TraceID: traceID}
	(*w).Write(resp.JSONBytes())
}

func WriteBadRequestDataResp(w *http.ResponseWriter, r *http.Request, msg string, v ...interface{}) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())
	resp := RespMsg{Code: http.StatusBadRequest, Msg: fmt.Sprintf(msg, v...), TraceID: traceID}
	(*w).Write(resp.JSONBytes())
}

func WriteInternalServerErrorResp(
	w *http.ResponseWriter, r *http.Request, err error, msg string, v ...interface{},
) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())

	if err != nil {
		msg += ":" + err.Error()
	}
	resp := RespMsg{
		Code:    http.StatusInternalServerError,
		Msg:     fmt.Sprintf(msg, v...),
		TraceID: traceID,
	}
	(*w).Write(resp.JSONBytes())
}

func WriteNeedLogin(w *http.ResponseWriter, r *http.Request) {
	resp := RespMsg{Code: http.StatusUnauthorized, Msg: "login needed"}
	(*w).Write(resp.JSONBytes())
}

func WriteNeedSuperUser(w *http.ResponseWriter, r *http.Request) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())
	resp := RespMsg{Code: http.StatusForbidden, Msg: "superuser needed", TraceID: traceID}
	(*w).Write(resp.JSONBytes())
}

func WritePermissionNotEnough(w *http.ResponseWriter, r *http.Request, msg string) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())
	resp := RespMsg{
		Code:    http.StatusForbidden,
		Msg:     "permission not enough:" + msg,
		TraceID: traceID,
	}
	(*w).Write(resp.JSONBytes())
}

func (resp *RespMsg) JSONBytes() []byte {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return r
}
