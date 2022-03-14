package client

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportFunctionExecuteHeartbeat(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "function execution heartbeat report"

	funcRunRecordID := ps.ByName("function_run_record_id")
	if funcRunRecordID == "" {
		scheduleLogger.Warningf(logTags, "function_run_record_id not in path")
		web.WriteBadRequestDataResp(&w, r, "function_run_record_id not in path")
		return
	}
	logTags["function_run_record_id"] = funcRunRecordID
	scheduleLogger.Infof(logTags, "received heartbeat report")

	funcRunRecordUUID, err := value_object.ParseToUUID(funcRunRecordID)
	if err != nil {
		scheduleLogger.Warningf(logTags, "parse to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r,
			"parse function_run_record_id to uuid failed: %v", err)
		return
	}

	err = heartbeatService.HeartBeatRepo.AliveReportByFuncRunRecordID(funcRunRecordUUID)
	if err != nil {
		scheduleLogger.Errorf(logTags, "heartbeat alive report failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "alive report failed")
	}

	scheduleLogger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}
