package flow_run_record

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

func Filter(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	filter, filterOption, err := BuildFromWebRequestParams(r.URL.Query())
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	flowRunRecord, err := fFRService.FlowRunRecord.Filter(*filter, *filterOption)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit repository failed")
		return
	}

	count, err := fFRService.FlowRunRecord.Count(*filter)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit total failed")
		return
	}

	web.WriteSucRespAndTotal(&w, fromAggSlice(flowRunRecord), count)
}
