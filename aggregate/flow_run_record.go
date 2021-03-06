package aggregate

import (
	"context"
	"fmt"
	"time"

	"github.com/fBloc/bloc-server/value_object"
)

type FlowRunRecord struct {
	ID                           value_object.UUID
	ArrangementID                value_object.UUID
	ArrangementFlowID            string
	ArrangementRunRecordID       string
	FlowID                       value_object.UUID
	FlowOriginID                 value_object.UUID
	FlowVersion                  uint
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
	InterceptMsg                 string
	RetriedAmount                uint16
	TimeoutCanceled              bool
	Canceled                     bool
	CancelUserID                 value_object.UUID
	TraceID                      string
	OverideIptParams             map[string][][]interface{}
}

func newFromFlow(ctx context.Context, f *Flow) *FlowRunRecord {
	ret := &FlowRunRecord{
		ID:            value_object.NewUUID(),
		FlowID:        f.ID,
		FlowOriginID:  f.OriginID,
		FlowVersion:   f.Version,
		TriggerSource: value_object.FlowTriggerSource,
		TriggerTime:   time.Now(),
		Status:        value_object.Created,
		TraceID:       value_object.GetTraceIDFromContext(ctx),
	}

	return ret
}

func NewKeyTriggeredFlowRunRecordWithCustomParam(
	ctx context.Context, f *Flow, key string,
	params map[string][][]interface{},
) (*FlowRunRecord, error) {
	rR := newFromFlow(ctx, f)
	rR.TriggerKey = key
	rR.TriggerType = value_object.Key
	rR.OverideIptParams = params
	return rR, nil
}

func NewKeyTriggeredFlowRunRecord(
	ctx context.Context, f *Flow, key string,
) (*FlowRunRecord, error) {
	rR := newFromFlow(ctx, f)
	rR.TriggerKey = key
	rR.TriggerType = value_object.Key
	return rR, nil
}

func NewUserTriggeredFlowRunRecord(
	ctx context.Context, f *Flow, triggerUser *User,
) (*FlowRunRecord, error) {
	if !f.UserCanExecute(triggerUser) {
		return nil, fmt.Errorf(
			"user: %s have no permission to trigger this flow",
			triggerUser.Name)
	}
	rR := newFromFlow(ctx, f)
	rR.TriggerUserID = triggerUser.ID
	rR.TriggerType = value_object.Manual
	return rR, nil
}

func NewCrontabTriggeredRunRecord(ctx context.Context, f *Flow) *FlowRunRecord {
	rR := newFromFlow(ctx, f)
	rR.TriggerType = value_object.Crontab
	return rR
}

func (task *FlowRunRecord) IsZero() bool {
	if task == nil {
		return true
	}
	return task.ID.IsNil()
}

func (task *FlowRunRecord) Finished() bool {
	return !task.EndTime.IsZero()
}
