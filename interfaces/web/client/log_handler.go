package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-backend-go/infrastructure/log"
	"github.com/fBloc/bloc-backend-go/interfaces/web"
	"github.com/fBloc/bloc-backend-go/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req LogHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	funcRunRecordUUID, err := value_object.ParseToUUID(req.FuncRunRecordID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse function_run_record_id to uuid failed: %s", err.Error())
		return
	}

	logger := log.New(
		value_object.FuncRunRecordLog.String()+"-"+funcRunRecordUUID.String(),
		logBackEnd)
	logger.ForceUpload()

	web.WritePlainSucOkResp(&w)
}
