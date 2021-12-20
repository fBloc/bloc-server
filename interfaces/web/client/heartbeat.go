package client

import (
	function_execute_heartbeat_repository "github.com/fBloc/bloc-server/repository/function_execute_heartbeat"
)

var heartbeatRepo function_execute_heartbeat_repository.FunctionExecuteHeartbeatRepository

func InjectHeartbeatRepo(
	hBR function_execute_heartbeat_repository.FunctionExecuteHeartbeatRepository,
) {
	heartbeatRepo = hBR
}

type FunctionExecuteHeartBeatHttpReq struct {
	FunctionRunRecordID string `json:"function_run_record_id"`
}
