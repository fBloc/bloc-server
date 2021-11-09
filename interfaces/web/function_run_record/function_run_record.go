package function_run_record

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/internal/json_date"
	"github.com/fBloc/bloc-backend-go/services/function_run_record"

	"github.com/google/uuid"
)

var fRRService *function_run_record.FunctionRunRecordService

func InjectFunctionRunRecordService(
	fRRS *function_run_record.FunctionRunRecordService,
) {
	fRRService = fRRS
}

type FunctionRunRecord struct {
	ID                uuid.UUID                    `json:"id"`
	FlowID            uuid.UUID                    `json:"flow_id"`
	FlowOriginID      uuid.UUID                    `json:"flow_origin_id"`
	ArrangementFlowID string                       `json:"arrangement_flow_id"`
	FunctionID        uuid.UUID                    `json:"function_id"`
	FlowFunctionID    string                       `json:"flow_function_id"`
	FlowRunRecordID   uuid.UUID                    `json:"flow_run_record_id"`
	Start             json_date.JsonDate           `json:"start"`
	End               json_date.JsonDate           `json:"end"`
	Suc               bool                         `json:"suc"`
	Pass              bool                         `json:"pass"`
	Canceled          bool                         `json:"canceled,omitempty"`
	Description       string                       `json:"description"`
	ErrorMsg          string                       `json:"error_msg"`
	IptBriefAndOsskey [][]aggregate.IptBriefAndKey `json:"ipt"`
	Progress          float32                      `json:"progress"`
	ProgressMsg       []string                     `json:"progress_msg"`
	ProcessStages     []string                     `json:"process_stages"`
	ProcessStageIndex int                          `json:"process_stage_index,omitempty"`
}

func fromAgg(
	aggFRR *aggregate.FunctionRunRecord,
) *FunctionRunRecord {
	if aggFRR.IsZero() {
		return nil
	}
	retFlow := &FunctionRunRecord{
		ID:                aggFRR.ID,
		FlowID:            aggFRR.FlowID,
		FlowOriginID:      aggFRR.FlowOriginID,
		ArrangementFlowID: aggFRR.ArrangementFlowID,
		FunctionID:        aggFRR.FunctionID,
		FlowFunctionID:    aggFRR.FlowFunctionID,
		FlowRunRecordID:   aggFRR.FlowRunRecordID,
		Start:             json_date.New(aggFRR.Start),
		End:               json_date.New(aggFRR.End),
		Suc:               aggFRR.Suc,
		Pass:              aggFRR.Pass,
		Canceled:          aggFRR.Canceled,
		Description:       aggFRR.Description,
		ErrorMsg:          aggFRR.ErrorMsg,
		IptBriefAndOsskey: aggFRR.IptBriefAndObskey,
		Progress:          aggFRR.Progress,
		ProgressMsg:       aggFRR.ProgressMsg,
		ProcessStages:     aggFRR.ProcessStages,
		ProcessStageIndex: aggFRR.ProcessStageIndex,
	}
	return retFlow
}

func fromAggSlice(
	aggFRRSlice []*aggregate.FunctionRunRecord,
) []*FunctionRunRecord {
	if len(aggFRRSlice) <= 0 {
		return []*FunctionRunRecord{}
	}
	ret := make([]*FunctionRunRecord, len(aggFRRSlice))
	for i, j := range aggFRRSlice {
		ret[i] = fromAgg(j)
	}
	return ret
}
