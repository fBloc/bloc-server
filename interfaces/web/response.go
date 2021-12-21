package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RespMsg struct {
	Code int         `json:"status_code"`
	Msg  string      `json:"status_msg"`
	Data interface{} `json:"data"`
}

func WriteSucResp(w *http.ResponseWriter, data interface{}) {
	resp := RespMsg{
		Code: http.StatusOK,
		Data: data}
	(*w).Write(resp.JSONBytes())
}

func WriteDeleteSucResp(w *http.ResponseWriter, delAmount int64) {
	resp := RespMsg{
		Code: http.StatusOK,
		Data: map[string]int64{"delete_amount": delAmount},
	}
	(*w).Write(resp.JSONBytes())
}

func WritePlainSucOkResp(w *http.ResponseWriter) {
	resp := RespMsg{Code: http.StatusOK, Data: "ok"}
	(*w).Write(resp.JSONBytes())
}

func WriteBadRequestDataResp(w *http.ResponseWriter, msg string, v ...interface{}) {
	resp := RespMsg{Code: http.StatusBadRequest, Msg: fmt.Sprintf(msg, v...)}
	(*w).Write(resp.JSONBytes())
}

func WriteInternalServerErrorResp(w *http.ResponseWriter, err error, msg string) {
	if err != nil {
		msg += ":" + err.Error()
	}
	resp := RespMsg{Code: http.StatusInternalServerError, Msg: msg}
	(*w).Write(resp.JSONBytes())
}

func WriteNeedLogin(w *http.ResponseWriter) {
	resp := RespMsg{Code: http.StatusUnauthorized, Msg: "login needed"}
	(*w).Write(resp.JSONBytes())
}

func WriteNeedSuperUser(w *http.ResponseWriter) {
	resp := RespMsg{Code: http.StatusForbidden, Msg: "superuser needed"}
	(*w).Write(resp.JSONBytes())
}

func WritePermissionNotEnough(w *http.ResponseWriter, msg string) {
	resp := RespMsg{Code: http.StatusForbidden,
		Msg: "permission not enough:" + msg}
	(*w).Write(resp.JSONBytes())
}

func (resp *RespMsg) JSONBytes() []byte {
	r, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
	}
	return r
}
