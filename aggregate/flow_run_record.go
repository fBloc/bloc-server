package aggregate

import (
	"time"

	"github.com/fBloc/bloc-backend-go/value_object"
)

type FlowRunRecord struct {
	ID                           value_object.UUID
	ArrangementID                value_object.UUID
	ArrangementFlowID            string
	ArrangementRunRecordID       string
	FlowID                       value_object.UUID
	FlowOriginID                 value_object.UUID
	FlowFuncIDMapFuncRunRecordID map[string]value_object.UUID
	TriggerType                  value_object.TriggerType
	TriggerKey                   string
	TriggerSource                value_object.FlowTriggeredSourceType
	TriggerUserID                value_object.UUID
	TriggerTime                  time.Time
	StartTime                    time.Time
	EndTime                      time.Time
	Status                       value_object.RunState
	ErrorMsg                     string
	RetriedAmount                uint16
	TimeoutCanceled              bool
	Canceled                     bool
	CancelUserID                 value_object.UUID
}

func newFromFlow(f Flow) *FlowRunRecord {
	ret := &FlowRunRecord{
		ID:            value_object.NewUUID(),
		FlowID:        f.ID,
		FlowOriginID:  f.OriginID,
		TriggerSource: value_object.FlowTriggerSource,
		TriggerTime:   time.Now(),
		Status:        value_object.Created,
	}

	return ret
}

func NewUserTriggeredRunRecord(f Flow, triggerUserID value_object.UUID) *FlowRunRecord {
	rR := newFromFlow(f)
	rR.TriggerUserID = triggerUserID
	rR.TriggerType = value_object.Manual
	return rR
}

func NewCrontabTriggeredRunRecord(f Flow) *FlowRunRecord {
	rR := newFromFlow(f)
	rR.TriggerType = value_object.Crontab
	return rR
}

func (task *FlowRunRecord) IsZero() bool {
	if task == nil {
		return true
	}
	return task.ID.IsNil()
}

func (task *FlowRunRecord) IsFromArrangement() bool {
	if task == nil {
		return false
	}
	if task.ArrangementID.IsNil() {
		return false
	}
	return true
}
