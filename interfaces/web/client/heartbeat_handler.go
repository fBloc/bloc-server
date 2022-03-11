package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportFunctionExecuteHeartbeat(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "function run alive report"

	var req FunctionExecuteHeartBeatHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		scheduleLogger.Warningf(logTags, "unmarshal body failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}
	logTags["function_run_record_id"] = req.FunctionRunRecordID

	funcRunRecordUUID, err := value_object.ParseToUUID(req.FunctionRunRecordID)
	if err != nil {
		scheduleLogger.Warningf(logTags, "parse to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse function_id to uuid failed: %v", err)
		return
	}

	fRRIns, err := fRRService.FunctionRunRecords.GetByID(funcRunRecordUUID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "get function_run_record failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "find function_run_record_ins by it's id failed")
		return
	}
	if fRRIns.IsZero() {
		scheduleLogger.Warningf(logTags, "get function_run_record by id match no record")
		web.WriteBadRequestDataResp(&w, r, "find no function_run_record_ins by this function_id")
		return
	}

	heartBeatIns, err := heartbeatService.HeartBeatRepo.GetByFunctionRunRecordID(funcRunRecordUUID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "get heartbeat by function_run_record_id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "no such function run record")
		return
	}
	logTags["heartbeat_id"] = heartBeatIns.ID.String()

	err = heartbeatService.HeartBeatRepo.AliveReport(heartBeatIns.ID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "heartbeat alive report failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "alive report failed")
	}

	scheduleLogger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}
