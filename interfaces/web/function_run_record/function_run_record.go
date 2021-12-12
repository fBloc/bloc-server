package function_run_record

import (
	"time"

	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/infrastructure/log_collect_backend"
	"github.com/fBloc/bloc-backend-go/internal/json_date"
	"github.com/fBloc/bloc-backend-go/services/function_run_record"
	"github.com/fBloc/bloc-backend-go/value_object"
	"github.com/spf13/cast"
)

var logBackend log_collect_backend.LogBackEnd
var fRRService *function_run_record.FunctionRunRecordService

func InjectFunctionRunRecordService(
	fRRS *function_run_record.FunctionRunRecordService,
) {
	fRRService = fRRS
}

func InjectLogCollectBackend(l log_collect_backend.LogBackEnd) {
	logBackend = l
}

type briefAndKey struct {
	IsArray          bool   `json:"is_array"`
	ValueType        string `json:"value_type"`
	Brief            string `json:"brief"`
	ObjectStorageKey string `json:"object_storage_key"`
}

type FunctionRunRecord struct {
	ID                          value_object.UUID      `json:"id"`
	FlowID                      value_object.UUID      `json:"flow_id"`
	FlowOriginID                value_object.UUID      `json:"flow_origin_id"`
	ArrangementFlowID           string                 `json:"arrangement_flow_id"`
	FunctionID                  value_object.UUID      `json:"function_id"`
	FlowFunctionID              string                 `json:"flow_function_id"`
	FlowRunRecordID             value_object.UUID      `json:"flow_run_record_id"`
	Start                       json_date.JsonDate     `json:"start"`
	End                         json_date.JsonDate     `json:"end"`
	Suc                         bool                   `json:"suc"`
	InterceptBelowFunctionRun   bool                   `json:"intercept_below_function_run"`
	Canceled                    bool                   `json:"canceled,omitempty"`
	Description                 string                 `json:"description"`
	ErrorMsg                    string                 `json:"error_msg"`
	IptBriefAndObjectStoragekey [][]briefAndKey        `json:"ipt"`
	OptBriefAndObjectStoragekey map[string]briefAndKey `json:"opt"`
	Progress                    float32                `json:"progress"`
	ProgressMsg                 []string               `json:"progress_msg"`
	ProcessStages               []string               `json:"process_stages"`
	ProcessStageIndex           int                    `json:"process_stage_index,omitempty"`
	FunctionProviderName        string                 `json:"function_provider_name"`
	ShouldBeCanceledAt          time.Time              `json:"should_be_canceled_at,omitempty"`
}

func fromAgg(
	aggFRR *aggregate.FunctionRunRecord,
) *FunctionRunRecord {
	if aggFRR.IsZero() {
		return nil
	}
	opt := make(map[string]briefAndKey, len(aggFRR.Opt))
	for k, v := range aggFRR.Opt {
		opt[k] = briefAndKey{
			IsArray:          aggFRR.OptKeyMapIsArray[k],
			ValueType:        string(aggFRR.OptKeyMapValueType[k]),
			Brief:            aggFRR.OptBrief[k],
			ObjectStorageKey: cast.ToString(v)}
	}

	ipt := make([][]briefAndKey, len(aggFRR.IptBriefAndObskey))
	for iptIndex, iptComponents := range aggFRR.IptBriefAndObskey {
		ipt[iptIndex] = make([]briefAndKey, len(iptComponents))
		for componentIndex, component := range iptComponents {
			ipt[iptIndex][componentIndex] = briefAndKey{
				IsArray:          component.IsArray,
				ValueType:        string(component.ValueType),
				Brief:            cast.ToString(component.Brief),
				ObjectStorageKey: component.FullKey}
		}
	}

	retFlow := &FunctionRunRecord{
		ID:                          aggFRR.ID,
		FlowID:                      aggFRR.FlowID,
		FlowOriginID:                aggFRR.FlowOriginID,
		ArrangementFlowID:           aggFRR.ArrangementFlowID,
		FunctionID:                  aggFRR.FunctionID,
		FlowFunctionID:              aggFRR.FlowFunctionID,
		FlowRunRecordID:             aggFRR.FlowRunRecordID,
		Start:                       json_date.New(aggFRR.Start),
		End:                         json_date.New(aggFRR.End),
		Suc:                         aggFRR.Suc,
		InterceptBelowFunctionRun:   aggFRR.InterceptBelowFunctionRun,
		Canceled:                    aggFRR.Canceled,
		Description:                 aggFRR.Description,
		ErrorMsg:                    aggFRR.ErrorMsg,
		IptBriefAndObjectStoragekey: ipt,
		OptBriefAndObjectStoragekey: opt,
		Progress:                    aggFRR.Progress,
		ProgressMsg:                 aggFRR.ProgressMsg,
		ProcessStages:               aggFRR.ProcessStages,
		ProcessStageIndex:           aggFRR.ProcessStageIndex,
		FunctionProviderName:        aggFRR.FunctionProviderName,
		ShouldBeCanceledAt:          aggFRR.ShouldBeCanceledAt,
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
