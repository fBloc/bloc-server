package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportProgress(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "report function run progress"

	var req ProgressReportHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		scheduleLogger.Warningf(logTags, "unmarshal body failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	funcRunRecordUUID, err := value_object.ParseToUUID(req.FunctionRunRecordID)
	if err != nil {
		msg := fmt.Sprintf("parse function_id to uuid failed: %v", err)
		scheduleLogger.Warningf(logTags, msg)
		web.WriteBadRequestDataResp(&w, r, msg)
		return
	}
	logTags["function_id"] = req.FunctionRunRecordID

	fRRIns, err := fRRService.FunctionRunRecords.GetByID(funcRunRecordUUID)
	if err != nil {
		scheduleLogger.Errorf(logTags,
			"find function_run_record_ins by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "find function_run_record_ins by id failed")
		return
	}
	if fRRIns.IsZero() {
		scheduleLogger.Warningf(logTags, "find no record")
		web.WriteBadRequestDataResp(&w, r, "find no function_run_record_ins by this function_id")
		return
	}

	if req.FuncRunProgress.Progress > 0 {
		scheduleLogger.Infof(logTags,
			"progress from %g to %g", fRRIns.Progress, req.FuncRunProgress.Progress)
		fRRService.FunctionRunRecords.PatchProgress(
			funcRunRecordUUID, req.FuncRunProgress.Progress)
	}
	if req.FuncRunProgress.Msg != "" {
		scheduleLogger.Infof(logTags,
			"add progress_msg: %s", req.FuncRunProgress.Msg)
		fRRService.FunctionRunRecords.PatchProgressMsg(
			funcRunRecordUUID, req.FuncRunProgress.Msg)
	}
	if req.FuncRunProgress.ProgressMilestoneIndex != nil {
		scheduleLogger.Infof(logTags,
			"progress index from %d to %d",
			fRRIns.ProgressMilestoneIndex, req.FuncRunProgress.ProgressMilestoneIndex)
		fRRService.FunctionRunRecords.PatchMilestoneIndex(
			funcRunRecordUUID, req.FuncRunProgress.ProgressMilestoneIndex)
	}

	scheduleLogger.Infof(logTags, "finished")
	web.WritePlainSucOkResp(&w, r)
}
