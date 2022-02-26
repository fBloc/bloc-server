package log_data

import (
	"time"

	"github.com/fBloc/bloc-server/infrastructure/log"
	"github.com/fBloc/bloc-server/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-server/value_object"
)

var (
	logBackend log_collect_backend.LogBackEnd
	logger     *log.Logger
)

type Req struct {
	LogType             value_object.LogType `json:"log_type"`
	StartTime           time.Time            `json:"start_time"`
	EndTime             time.Time            `json:"end_time"`
	FunctionRunRecordID value_object.UUID    `json:"function_run_record_id"`
}

func InjectLogCollectBackend(l log_collect_backend.LogBackEnd) {
	logBackend = l
}

func InjectLogger(
	l *log.Logger,
) {
	logger = l
}
