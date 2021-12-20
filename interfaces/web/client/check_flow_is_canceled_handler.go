package client

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func FlowRunRecordIsCanceled(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	flowRunRecordIDStr := ps.ByName("id")
	if flowRunRecordIDStr == "" {
		web.WriteBadRequestDataResp(&w, "flow_run_record_id cannot be blank")
		return
	}

	flowRunRecordUUID, err := value_object.ParseToUUID(flowRunRecordIDStr)
	if err != nil {
		web.WriteBadRequestDataResp(
			&w, "flow_run_record_id parse to uuid failed: %s",
			err.Error())
		return
	}

	flowRunRecordIns, err := flowRunRecordService.FlowRunRecord.GetByID(flowRunRecordUUID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit flow_run_record_ins by id failed")
		return
	}

	web.WriteSucResp(&w, map[string]bool{"canceled": flowRunRecordIns.Canceled})
}
