package client

import (
	"time"

	"github.com/fBloc/bloc-server/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-server/value_object"
)

var logBackEnd log_collect_backend.LogBackEnd

func InjectLogBackend(lB log_collect_backend.LogBackEnd) {
	logBackEnd = lB
}

type msg struct {
	Level  value_object.LogLevel `json:"level"`
	TagMap map[string]string     `json:"tag_map"`
	Data   string                `json:"data"`
	Time   time.Time             `json:"time"`
}

type FuncRunLogHttpReq struct {
	LogData []*msg `json:"logs"`
}
