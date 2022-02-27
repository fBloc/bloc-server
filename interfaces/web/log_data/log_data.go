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
	LogType    value_object.LogType `json:"log_type"`
	TagFilters map[string]string    `json:"tag_filter"`
	StartTime  time.Time            `json:"start_time"`
	EndTime    time.Time            `json:"end_time"`
}

func InjectLogCollectBackend(l log_collect_backend.LogBackEnd) {
	logBackend = l
}

func InjectLogger(
	l *log.Logger,
) {
	logger = l
}
