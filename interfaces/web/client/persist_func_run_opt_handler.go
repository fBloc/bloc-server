package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func PersistFuncRunOptField(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "persist function run opt field"

	var req PersistFuncRunOptFieldHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		scheduleLogger.Warningf(logTags, "unmarshal body failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}
	logTags["function_run_record_id"] = req.FunctionRunRecordID.String()
	logTags["opt_key"] = req.OptKey

	amount, err := fRRService.FunctionRunRecords.Count(
		*value_object.NewRepositoryFilter().AddEqual("id", req.FunctionRunRecordID))
	if err != nil {
		scheduleLogger.Errorf(logTags, "find function_run_record failed: %v", err)
		web.WriteInternalServerErrorResp(
			&w, r, err, "find corresponding function_run_record error")
		return
	}
	if amount != 1 {
		scheduleLogger.Warningf(logTags, "find no function_run_record record")
		web.WriteBadRequestDataResp(
			&w, r, "this function_run_record_id find no record")
		return
	}

	uploadByte, _ := json.Marshal(req.Data)
	ossKey := req.FunctionRunRecordID.String() + "_" + req.OptKey
	err = objectStorage.Set(ossKey, uploadByte)
	if err != nil {
		scheduleLogger.Errorf(logTags,
			"save to object storage failed: %v", err)
		web.WriteInternalServerErrorResp(
			&w, r, err, "save to object storage failed")
		return
	}
	logTags["object_storage_key"] = ossKey

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
	logTags["brief"] = resp.Brief

	scheduleLogger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, resp)
}
