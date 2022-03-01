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
	resp := RespMsg{Code: http.StatusOK, Data: data}
	resp.writeResp(w, r)
}

func WriteDeleteSucResp(w *http.ResponseWriter, r *http.Request, delAmount int64) {
	resp := RespMsg{
		Code: http.StatusOK,
		Data: map[string]int64{"delete_amount": delAmount}}
	resp.writeResp(w, r)
}

func WritePlainSucOkResp(w *http.ResponseWriter, r *http.Request) {
	resp := RespMsg{Code: http.StatusOK, Data: "ok"}
	resp.writeResp(w, r)
}

func WriteBadRequestDataResp(w *http.ResponseWriter, r *http.Request, msg string, v ...interface{}) {
	resp := RespMsg{Code: http.StatusBadRequest, Msg: fmt.Sprintf(msg, v...)}
	resp.writeResp(w, r)
}

func WriteInternalServerErrorResp(
	w *http.ResponseWriter, r *http.Request, err error, msg string, v ...interface{},
) {
	if err != nil {
		msg += ":" + err.Error()
	}
	resp := RespMsg{
		Code: http.StatusInternalServerError,
		Msg:  fmt.Sprintf(msg, v...),
	}
	resp.writeResp(w, r)
}

func WriteNeedLogin(w *http.ResponseWriter, r *http.Request) {
	resp := RespMsg{Code: http.StatusUnauthorized, Msg: "login needed"}
	resp.writeResp(w, r)
}

func WriteNeedSuperUser(w *http.ResponseWriter, r *http.Request) {
	resp := RespMsg{Code: http.StatusForbidden, Msg: "superuser needed"}
	resp.writeResp(w, r)
}

func WritePermissionNotEnough(w *http.ResponseWriter, r *http.Request, msg string) {
	resp := RespMsg{
		Code: http.StatusForbidden,
		Msg:  "permission not enough:" + msg,
	}
	resp.writeResp(w, r)
}

func (resp *RespMsg) jSONBytes() []byte {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return r
}

func (resp *RespMsg) writeResp(w *http.ResponseWriter, r *http.Request) {
	traceID, _ := GetReqTraceIDFromContext(r.Context())
	resp.TraceID = traceID
	(*w).Header().Add("content-type", "application/json")
	(*w).Write(resp.jSONBytes())
}
