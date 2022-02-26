package flow_run_record

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"

	"github.com/julienschmidt/httprouter"
)

func Filter(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "filter flow_run_record"

	filter, filterOption, err := BuildFromWebRequestParams(r.URL.Query())
	if err != nil {
		fFRService.Logger.Warningf(
			logTags, "parse filter options from get param failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	flowRunRecord, err := fFRService.FlowRunRecord.Filter(*filter, *filterOption)
	if err != nil {
		fFRService.Logger.Errorf(logTags, "filter failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit repository failed")
		return
	}

	count, err := fFRService.FlowRunRecord.Count(*filter)
	if err != nil {
		fFRService.Logger.Errorf(logTags, "get count failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit total failed")
		return
	}

	resp := FlowRunRecordFilterResp{
		Total: count,
		Items: fromAggSlice(flowRunRecord),
	}

	fFRService.Logger.Infof(logTags, "finished with amount: %d", count)
	web.WriteSucResp(&w, r, resp)
}
