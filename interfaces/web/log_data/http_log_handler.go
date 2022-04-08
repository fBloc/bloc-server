package log_data

import (
	"encoding/json"
	"net/http"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

func PullLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logTags := web.GetTraceAboutFields(r.Context())
	logTags["business"] = "pull log"

	var req Req
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Warningf(logTags,
			"json unmarshal req body failed: %v", err)
		web.WriteBadRequestDataResp(&w, r, err.Error())
		return
	}

	var thisLog *log.Logger = nil
	if req.LogType == value_object.HttpServerLog {
		thisLog = log.New(value_object.HttpServerLog.String(), logBackend)
	} else if req.LogType == value_object.ScheduleLog {
		thisLog = log.New(value_object.ScheduleLog.String(), logBackend)
	} else if req.LogType == value_object.FuncRunRecordLog {
		msg := "this api not provide function_run_record log search"
		logger.Warningf(logTags, msg)
		web.WriteBadRequestDataResp(&w, r, msg)
		return
	} else {
		// when logType is blank, it means only use tag_filter to filter out log
		// typical use case is get certain trace_id's log. {"trace_id": "xxx"} ...
		thisLog = log.New("", logBackend)
	}

	if thisLog == nil {
		logger.Errorf(logTags, "get logger failed")
		web.WriteInternalServerErrorResp(&w, r, errors.New("get logger failed"), "")
		return
	}

	logStrSlice, err := thisLog.PullLogBetweenTime(
		req.TagFilters, req.StartTime.ToTime(), req.EndTime.ToTime())
	if err != nil {
		logger.Errorf(logTags, "pull log failed: %v", err)
		web.WriteInternalServerErrorResp(&w, r, err, "PullLogBetweenTime failed")
		return
	}

	web.WriteSucResp(&w, r, logStrSlice)
}
