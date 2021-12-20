package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/julienschmidt/httprouter"
)

func PersistFuncRunOptField(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req PersistFuncRunOptFieldHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	// TODO 检测req.FunctionRunRecordID是否有效
	uploadByte, _ := json.Marshal(req.Data)
	ossKey := req.FunctionRunRecordID.String() + "_" + req.OptKey
	err = objectStorage.Set(ossKey, uploadByte)
	if err != nil {
		web.WriteInternalServerErrorResp(
			&w, err, "save to object storage failed")
		return
	}

	resp := PersistFuncRunOptFieldHttpResp{
		ObjectStorageKey: ossKey,
	}

	optInRune := []rune(string(uploadByte))
	// TODO 配置化截断长度
	minLength := 51
	if len(optInRune) < minLength {
		minLength = len(optInRune)
	}
	resp.Brief = string(optInRune[:minLength-1])

	web.WriteSucResp(&w, resp)
}
