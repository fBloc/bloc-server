package function_run_record

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

func GetProgressByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get function_run_record progress"

	id := ps.ByName("id")
	if id == "" {
		fRRService.Logger.Warningf(logTags, "miss id in url path")
		web.WriteBadRequestDataResp(&w, r, "id cannot be null")
		return
	}
	logTags["id"] = id

	uuID, err := value_object.ParseToUUID(id)
	if err != nil {
		fRRService.Logger.Warningf(logTags, "get by id failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse id to uuid failed")
		return
	}

	aggFRR, err := fRRService.FunctionRunRecords.GetOnlyProgressInfoByID(uuID)
	if err != nil {
		fRRService.Logger.Errorf(logTags, "get by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit repository failed")
		return
	}

	fRRService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, fromAggToProgress(aggFRR))
}

func Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "get function_run_record"

	id := ps.ByName("id")
	if id == "" {
		fRRService.Logger.Warningf(logTags, "miss id in url path")
		web.WriteBadRequestDataResp(&w, r, "id cannot be null")
		return
	}
	logTags["id"] = id

	uuID, err := value_object.ParseToUUID(id)
	if err != nil {
		fRRService.Logger.Warningf(logTags, "get by id failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse id to uuid failed")
		return
	}

	aggFRR, err := fRRService.FunctionRunRecords.GetByID(uuID)
	if err != nil {
		fRRService.Logger.Errorf(logTags, "get by id failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit repository failed")
		return
	}

	fRRService.Logger.Infof(logTags, "finished")
	web.WriteSucResp(&w, r, fromAgg(aggFRR))
}

func Filter(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "filter function_run_record"

	filter, filterOption, err := BuildFromWebRequestParams(r.URL.Query())
	if err != nil {
		fRRService.Logger.Errorf(
			logTags, "build filter from get param failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	aggFunctionRunRecords, err := fRRService.FunctionRunRecords.Filter(*filter, *filterOption)
	if err != nil {
		fRRService.Logger.Errorf(
			logTags, "filer function run record failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit repository failed")
		return
	}

	count, err := fRRService.FunctionRunRecords.Count(*filter)
	if err != nil {
		fRRService.Logger.Errorf(logTags, "count failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "visit total failed")
		return
	}

	resp := FunctionRunRecordFilterResp{
		Total: count,
		Items: fromAggSlice(aggFunctionRunRecords),
	}

	fRRService.Logger.Infof(logTags, "finished with amount: %d", count)
	web.WriteSucResp(&w, r, resp)
}
