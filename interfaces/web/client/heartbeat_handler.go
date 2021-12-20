package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportFunctionExecuteHeartbeat(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req FunctionExecuteHeartBeatHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	funcRunRecordUUID, err := value_object.ParseToUUID(req.FunctionRunRecordID)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse function_id to uuid failed:", err.Error())
		return
	}

	fRRIns, err := fRRService.FunctionRunRecords.GetByID(funcRunRecordUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "find function_run_record_ins by it's id failed")
		return
	}
	if fRRIns.IsZero() {
		web.WriteBadRequestDataResp(&w, "find no function_run_record_ins by this function_id")
		return
	}

	heartBeatIns, err := heartbeatRepo.GetByFunctionRunRecordID(funcRunRecordUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "no such function run record")
		return
	}
	err = heartbeatRepo.AliveReport(heartBeatIns.ID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "alive report failed")
	}
	web.WritePlainSucOkResp(&w)
}
