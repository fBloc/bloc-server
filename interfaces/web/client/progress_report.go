package client

type highReadableFunctionRunProgress struct {
	Progress          float32 `json:"progress"`
	Msg               string  `json:"msg"`
	ProcessStageIndex int     `json:"process_stage_index"`
}

type progressReportHttpReq struct {
	FunctionRunRecordID string                          `json:"function_run_record_id"`
	FuncRunProgress     highReadableFunctionRunProgress `json:"high_readable_run_progress"`
}
