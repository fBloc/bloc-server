package client

type HighReadableFunctionRunProgress struct {
	Progress          float32 `json:"progress"`
	Msg               string  `json:"msg"`
	ProcessStageIndex int     `json:"process_stage_index"`
}

type ProgressReportHttpReq struct {
	FunctionRunRecordID string                          `json:"function_run_record_id"`
	FuncRunProgress     HighReadableFunctionRunProgress `json:"high_readable_run_progress"`
}
