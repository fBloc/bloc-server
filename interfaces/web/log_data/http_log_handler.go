package log_data

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/interfaces/web"
	"github.com/fBloc/bloc-server/internal/json_date"
	"github.com/fBloc/bloc-server/value_object"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

func PullLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req Req
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		web.WriteBadRequestDataResp(&w, err.Error())
		return
	}
	if !req.LogType.IsValid() {
		web.WriteBadRequestDataResp(&w, "log_type not valid")
		return
	}
	if req.StartTime.IsZero() {
		web.WriteBadRequestDataResp(
			&w, "start_time cannot be nil")
		return
	}
	if req.EndTime.IsZero() {
		req.EndTime = json_date.Now()
	}

	var logger *log.Logger
	if req.LogType == value_object.HttpServerLog {
		logger = log.New(value_object.HttpServerLog.String(), logBackend)
	} else if req.LogType == value_object.ScheduleLog {
		logger = log.New(value_object.ScheduleLog.String(), logBackend)
	} else if req.LogType == value_object.FuncRunRecordLog {
		web.WriteBadRequestDataResp(&w, "this api not provide function_run_record log search")
		return
	}
	if logger == nil {
		web.WriteInternalServerErrorResp(&w, errors.New("get logger failed"), "")
		return
	}

	logStrSlice, err := logger.PullLogBetweenTime(
		map[string]string{}, req.StartTime.Time, req.EndTime.Time)
	if err != nil {
		web.WriteInternalServerErrorResp(&w, err, "PullLogBetweenTime failed")
		return
	}

	web.WriteSucResp(&w, logStrSlice)
}

func PullFunctionRunRecordLog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	functionRunRecordIDStr := ps.ByName("function_run_record_id")
	_, err := value_object.ParseToUUID(functionRunRecordIDStr)
	if err != nil {
		web.WriteBadRequestDataResp(&w, "parse function_run_record_id to uuid failed. error: %+v", err)
		return
	}

	// 如果是ing的任务，前端可能需要一直拉取最新的数据。
	// 而上次获取的时候已经拿了一部分了，所以就从上次的数据末尾继续获取就是了
	var start time.Time
	startTimeStr := r.URL.Query().Get("start")
	if startTimeStr != "" {
		start, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			web.WriteBadRequestDataResp(&w, "parse start failed. error: %+v", err)
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
		web.WriteInternalServerErrorResp(
			&w, err, "logger.PullLogBetweenTime error: %v", err)
		return
	}
	web.WriteSucResp(&w, logs)
}
