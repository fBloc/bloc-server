package client

import (
	"time"

	"github.com/fBloc/bloc-backend-go/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-backend-go/value_object"
)

var logBackEnd log_collect_backend.LogBackEnd

func InjectLogBackend(lB log_collect_backend.LogBackEnd) {
	logBackEnd = lB
}

type msg struct {
	Level value_object.LogLevel `json:"level"`
	Data  string                `json:"data"`
	Time  time.Time             `json:"time"`
}

type LogHttpReq struct {
	Name    string `json:"name"`
	LogData []*msg `json:"log_data"`
}
