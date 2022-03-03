package client

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
)

func ReportLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO 1. 需要显示的指定TagMap中应包含function_run_record_id （同步client也要修改）；2. 补上日志
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "report function run log"

	var req FuncRunLogHttpReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	thisLog := log.New(
		value_object.FuncRunRecordLog.String(),
		logBackEnd)
	// 这里实际只是帮助client记录了client提交的日志，故trace_id和span_id都应该直接用client的
	// 而client是通过在header中携带这两个信息的
	clientTraceID, _ := web.GetReqTraceIDFromContext(r.Context())
	clientSpanID, _ := web.GetReqParentSpanIDFromContext(r.Context())
	for _, logItem := range req.LogData {
		if logItem.TagMap == nil {
			logItem.TagMap = make(map[string]string)
		}
		logItem.TagMap["business"] = "client's log"
		logItem.TagMap[string(value_object.TraceID)] = clientTraceID
		logItem.TagMap[string(value_object.SpanID)] = clientSpanID
		thisLog.UploadLog(
			logItem.Time, logItem.Level,
			logItem.TagMap, logItem.Data)
	}
	thisLog.ForceUpload()

	web.WritePlainSucOkResp(&w, r)
}
