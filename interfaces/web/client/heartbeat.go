package client

import "github.com/fBloc/bloc-server/services/function_execute_heartbeat"

var heartbeatService *function_execute_heartbeat.FunctionExecuteHeartbeatService

func InjectHeartbeatService(
	hBR *function_execute_heartbeat.FunctionExecuteHeartbeatService,
) {
	heartbeatService = hBR
}

type FunctionExecuteHeartBeatHttpReq struct {
	FunctionRunRecordID string `json:"function_run_record_id"`
}
