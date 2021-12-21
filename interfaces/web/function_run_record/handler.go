package function_run_record

import (
	"net/http"

	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

func Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if id == "" {
		web.WriteBadRequestDataResp(&w, "id cannot be null")
		return
	}
	uuID, err := value_object.ParseToUUID(id)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse id to uuid failed")
		return
	}

	aggFRR, err := fRRService.FunctionRunRecords.GetByID(uuID)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit repository failed")
		return
	}
	web.WriteSucResp(&w, fromAgg(aggFRR))
}

func Filter(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	filter, filterOption, err := BuildFromWebRequestParams(r.URL.Query())
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}

	aggFunctionRunRecords, err := fRRService.FunctionRunRecords.Filter(*filter, *filterOption)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit repository failed")
		return
	}

	count, err := fRRService.FunctionRunRecords.Count(*filter)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "visit total failed")
		return
	}

	resp := FunctionRunRecordFilterResp{
		Total: count,
		Items: fromAggSlice(aggFunctionRunRecords),
	}

	web.WriteSucResp(&w, resp)
}
