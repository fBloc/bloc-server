package aggregate

import (
	"time"

	"github.com/fBloc/bloc-server/pkg/value_type"
	"github.com/fBloc/bloc-server/value_object"
)

type IptBriefAndKey struct {
	IsArray   bool
	ValueType value_type.ValueType
	Brief     interface{}
	FullKey   string
}

type FunctionRunRecord struct {
	ID                        value_object.UUID
	FlowID                    value_object.UUID
	FlowOriginID              value_object.UUID
	ArrangementFlowID         string
	FunctionID                value_object.UUID
	FlowFunctionID            string
	FlowRunRecordID           value_object.UUID
	FunctionProviderName      string
	ShouldBeCanceledAt        time.Time
	Start                     time.Time
	End                       time.Time
	Suc                       bool
	InterceptBelowFunctionRun bool
	Canceled                  bool
	Description               string
	ErrorMsg                  string
	Ipts                      [][]interface{}    // 实际调用user implement function的时候，根据用户前端输入的匹配规则/值，获取到实际值填充进此字段，作为参数传递给函数
	IptBriefAndObskey         [][]IptBriefAndKey // Ipts中的值可能非常大，在前端显示的时候不可能全部显示，故进行截断，将真实值保存到对象存储，需要的时候才查看真实值
	Opt                       map[string]interface{}
	OptBrief                  map[string]string
	OptKeyMapValueType        map[string]value_type.ValueType
	OptKeyMapIsArray          map[string]bool
	Progress                  float32
	ProgressMsg               []string
	ProcessStages             []string
	ProcessStageIndex         int
}

func NewFunctionRunRecordFromFlowDriven(
	functionIns Function,
	flowRunRecordIns FlowRunRecord,
	flowFunctionID string,
) *FunctionRunRecord {
	return &FunctionRunRecord{
		ID:                   value_object.NewUUID(),
		FlowID:               flowRunRecordIns.FlowID,
		FlowOriginID:         flowRunRecordIns.FlowOriginID,
		FunctionID:           functionIns.ID,
		FlowFunctionID:       flowFunctionID,
		FlowRunRecordID:      flowRunRecordIns.ID,
		Start:                time.Now(),
		ProcessStages:        functionIns.ProcessStages,
		FunctionProviderName: functionIns.ProviderName,
	}
}

func (bh *FunctionRunRecord) IsZero() bool {
	if bh == nil {
		return true
	}
	return bh.ID.IsNil()
}

func (bh *FunctionRunRecord) UsedSeconds() float64 {
	if bh.IsZero() {
		return 0
	}
	end := bh.End
	if end.IsZero() {
		end = time.Now()
	}
	return end.Sub(bh.Start).Seconds()
}

func (bh *FunctionRunRecord) Failed() bool {
	if bh.IsZero() {
		return false
	}
	if bh.End.IsZero() {
		return false
	}
	return !bh.Suc
}

func (bh *FunctionRunRecord) Finished() bool {
	if bh.IsZero() {
		return false
	}
	if bh.End.IsZero() {
		return false
	}
	return true
}

func (bh *FunctionRunRecord) SetSuc() {
	if bh.IsZero() {
		return
	}
	bh.Suc = true
	bh.End = time.Now()
}

func (bh *FunctionRunRecord) SetFail(errorMsg string) {
	if bh.IsZero() {
		return
	}
	bh.Suc = false
	bh.ErrorMsg = errorMsg
	bh.End = time.Now()
}
