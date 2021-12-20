package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportProgress(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req progressReportHttpReq
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

	if req.FuncRunProgress.Progress > 0 {
		fRRService.FunctionRunRecords.PatchProgress(
			funcRunRecordUUID, req.FuncRunProgress.Progress)
	}
	if req.FuncRunProgress.Msg != "" {
		fRRService.FunctionRunRecords.PatchProgressMsg(
			funcRunRecordUUID, req.FuncRunProgress.Msg)
	}
	if req.FuncRunProgress.ProcessStageIndex > 0 {
		fRRService.FunctionRunRecords.PatchStageIndex(
			funcRunRecordUUID, req.FuncRunProgress.ProcessStageIndex)
	}

	web.WritePlainSucOkResp(&w)
}
