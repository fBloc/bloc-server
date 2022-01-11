package log_data

import (
	"encoding/json"
	"net/http"

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
	} else if req.LogType == value_object.ConsumerLog {
		logger = log.New(value_object.ConsumerLog.String(), logBackend)
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
