package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func PersistFuncRunOptField(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req PersistFuncRunOptFieldHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}
	fRRService.Logger.Infof(
		map[string]string{"function_run_record_id": req.FunctionRunRecordID.String()},
		"received persist function run opt filed.function_run_record_id: %s",
		req.FunctionRunRecordID)

	amount, err := fRRService.FunctionRunRecords.Count(
		*value_object.NewRepositoryFilter().AddEqual("id", req.FunctionRunRecordID))
	if err != nil {
		web.WriteInternalServerErrorResp(
			&w, err, "find corresponding function_run_record error")
		return
	}
	if amount != 1 {
		web.WriteBadRequestDataResp(
			&w, "this function_run_record_id find no record")
		return
	}

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
	minLength := req.BriefCutLength
	if minLength <= 0 {
		minLength = 51
	}

	if len(optInRune) < minLength {
		minLength = len(optInRune)
	}
	resp.Brief = string(optInRune[:minLength-1])
	fRRService.Logger.Infof(
		map[string]string{"function_run_record_id": req.FunctionRunRecordID.String()},
		`persist function run opt filed.function_run_record_id: %s, key: %s, brief_opt: %s, objectStorage_key:%s`,
		req.FunctionRunRecordID, req.OptKey,
		resp.Brief, resp.ObjectStorageKey)

	web.WriteSucResp(&w, resp)
}
