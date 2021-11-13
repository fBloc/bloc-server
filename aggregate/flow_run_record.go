package aggregate

import (
	"time"

	"github.com/fBloc/bloc-backend-go/value_object"

	"github.com/google/uuid"
)

type FlowRunRecord struct {
	ID                           uuid.UUID
	ArrangementID                uuid.UUID
	ArrangementFlowID            string
	ArrangementRunRecordID       string
	FlowID                       uuid.UUID
	FlowOriginID                 uuid.UUID
	FlowFuncIDMapFuncRunRecordID map[string]uuid.UUID
	TriggerType                  value_object.TriggerType
	TriggerKey                   string
	TriggerSource                value_object.FlowTriggeredSourceType
	TriggerUserID                uuid.UUID
	TriggerTime                  time.Time
	StartTime                    time.Time
	EndTime                      time.Time
	Status                       value_object.RunState
	ErrorMsg                     string
	RetriedAmount                uint16
	TimeoutCanceled              bool
	Canceled                     bool
	CancelUserID                 uuid.UUID
}

func newFromFlow(f Flow) *FlowRunRecord {
	ret := &FlowRunRecord{
		ID:            uuid.New(),
		FlowID:        f.ID,
		FlowOriginID:  f.OriginID,
		TriggerSource: value_object.FlowTriggerSource,
		TriggerTime:   time.Now(),
		Status:        value_object.Created,
	}

	return ret
}

func NewUserTriggeredRunRecord(f Flow, triggerUserID uuid.UUID) *FlowRunRecord {
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
	return task.ID == uuid.Nil
}

func (task *FlowRunRecord) IsFromArrangement() bool {
	if task == nil {
		return false
	}
	if task.ArrangementID == uuid.Nil {
		return false
	}
	return true
}
