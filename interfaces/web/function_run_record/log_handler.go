package function_run_record

import (
	"net/http"
	"time"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"

	"github.com/julienschmidt/httprouter"
)

func PullLog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "pull function_run_record's log"

	functionRunRecordIDStr := ps.ByName("function_run_record_id")
	_, err := value_object.ParseToUUID(functionRunRecordIDStr)
	if err != nil {
		fRRService.Logger.Warningf(logTags,
			"parse function_run_record_id to uuid failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, "parse function_run_record_id to uuid failed. error: %+v", err)
		return
	}
	logTags["id"] = functionRunRecordIDStr

	// 如果是ing的任务，前端可能需要一直拉取最新的数据。
	// 而上次获取的时候已经拿了一部分了，所以就从上次的数据末尾继续获取就是了
	var start time.Time
	startTimeStr := r.URL.Query().Get("start")
	if startTimeStr != "" {
		start, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			fRRService.Logger.Warningf(logTags,
				"parse start: %s in get param failed: %v", startTimeStr, err)
			web.WriteBadRequestDataResp(&w, r, "parse start failed. error: %+v", err)
			return
		}
	}

	logger := log.New(
		value_object.FuncRunRecordLog.String(),
		logBackend)
	logs, err := logger.PullLogBetweenTime(
		map[string]string{"function_run_record_id": functionRunRecordIDStr},
		start, time.Time{})
	if err != nil {
		fRRService.Logger.Errorf(logTags,
			"pull log failed: %v", startTimeStr, err)
		web.WriteInternalServerErrorResp(
			&w, r, err, "logger.PullLogBetweenTime error: %v", err)
		return
	}

	fRRService.Logger.Infof(logTags,
		"finished with amount: %d", len(logs))
	web.WriteSucResp(&w, r, logs)
}
