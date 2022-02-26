package client

import (
	"github.com/fBloc/bloc-server/infrastructure/log"

	"github.com/fBloc/bloc-server/services/flow"
	"github.com/fBloc/bloc-server/services/flow_run_record"
	"github.com/fBloc/bloc-server/services/function_run_record"
)

var fRRService *function_run_record.FunctionRunRecordService

func InjectFunctionRunRecordService(
	fRRS *function_run_record.FunctionRunRecordService,
) {
	fRRService = fRRS
}

var flowService *flow.FlowService

func InjectFlowService(
	f *flow.FlowService,
) {
	flowService = f
}

var flowRunRecordService *flow_run_record.FlowRunRecordService

func InjectFlowRunRecordService(
	f *flow_run_record.FlowRunRecordService,
) {
	flowRunRecordService = f
}

var scheduleLogger *log.Logger

func InjectScheduleLogger(
	l *log.Logger,
) {
	scheduleLogger = l
}

type FuncRunFinishedHttpReq struct {
	FunctionRunRecordID       string            `json:"function_run_record_id"`
	Suc                       bool              `json:"suc"`
	Canceled                  bool              `json:"canceled"`
	TimeoutCanceled           bool              `json:"timeout_canceled"`
	InterceptBelowFunctionRun bool              `json:"intercept_below_function_run"`
	ErrorMsg                  string            `json:"error_msg"`
	Description               string            `json:"description"`
	OptKeyMapBriefData        map[string]string `json:"optKey_map_briefData"`
	OptKeyMapObjectStorageKey map[string]string `json:"optKey_map_objectStorageKey"`
}
