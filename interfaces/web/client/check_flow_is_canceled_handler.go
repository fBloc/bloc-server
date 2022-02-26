package client

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func FlowRunRecordIsCanceled(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get whether flow_run_record is canceled"

	flowRunRecordIDStr := ps.ByName("id")
	if flowRunRecordIDStr == "" {
		fRRService.Logger.Warningf(logTags, "miss id in url path")
		web.WriteBadRequestDataResp(&w, r, "flow_run_record_id cannot be blank")
		return
	}
	logTags["flow_run_record_id"] = flowRunRecordIDStr

	flowRunRecordUUID, err := value_object.ParseToUUID(flowRunRecordIDStr)
	if err != nil {
		fRRService.Logger.Warningf(
			logTags, "parse to uuid failed: %v", err)
		web.WriteBadRequestDataResp(
			&w, r, "flow_run_record_id parse to uuid failed: %s",
			err.Error())
		return
	}

	flowRunRecordIns, err := flowRunRecordService.FlowRunRecord.GetByID(flowRunRecordUUID)
	if err != nil {
		fRRService.Logger.Errorf(
			logTags, "get by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit flow_run_record_ins by id failed")
		return
	}

	fRRService.Logger.Errorf(logTags,
		"finished with whether canceld: %t", flowRunRecordIns.Canceled)
	web.WriteSucResp(&w, r, map[string]bool{"canceled": flowRunRecordIns.Canceled})
}
