package function_run_record

import (
	"net/http"
	"time"

	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/julienschmidt/httprouter"
)

func ListLogKeys(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		web.WriteBadRequestDataResp(&w, "id cannot be null")
		return
	}
	uuID, err := value_object.ParseToUUID(id)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse id to uuid failed")
		return
	}

	aggFRR, err := fRRService.FunctionRunRecords.GetByID(uuID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit repository failed")
		return
	}

	startTime := aggFRR.Start
	endTime := aggFRR.End
	if endTime.IsZero() { // 如果没有endTime，表示还在运行中，先获取到当前就行
		endTime = time.Now()
	}
	keys, err := logBackend.ListKeysBetween(
		value_object.FuncRunRecordLog.String()+"-"+uuID.String(),
		startTime, endTime)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "list keys failed")
		return
	}

	web.WriteSucResp(&w, keys)
}

func GetLogByKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logKey := ps.ByName("log_key")
	if logKey == "" {
		web.WriteBadRequestDataResp(&w, "log_key cannot be null")
		return
	}

	data, err := logBackend.PullDataByKey(logKey)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "pull log by key failed")
		return
	}
	web.WriteSucResp(&w, data)
}
