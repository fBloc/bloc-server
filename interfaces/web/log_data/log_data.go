package log_data

import (
	"github.com/fBloc/bloc-backend-go/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-backend-go/internal/json_date"
	"github.com/fBloc/bloc-backend-go/value_object"
)

var logBackend log_collect_backend.LogBackEnd

type Req struct {
	LogType             value_object.LogType `json:"log_type"`
	StartTime           json_date.JsonDate   `json:"start_time"`
	EndTime             json_date.JsonDate   `json:"end_time"`
	FunctionRunRecordID value_object.UUID    `json:"function_run_record_id"`
}

func InjectLogCollectBackend(l log_collect_backend.LogBackEnd) {
	logBackend = l
}
