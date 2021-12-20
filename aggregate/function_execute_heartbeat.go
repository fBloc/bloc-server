package aggregate

import (
	"time"

	"github.com/fBloc/bloc-server/value_object"
)

const (
	HeartBeatReportInterval = 5 * time.Second
	HeartBeatDeadThreshold  = 30 * time.Second
)

type FunctionExecuteHeartBeat struct {
	ID                  value_object.UUID
	FunctionRunRecordID value_object.UUID
	StartTime           time.Time
	LatestHeartbeatTime time.Time
}

func NewFunctionExecuteHeartBeat(
	functionRunRecordID value_object.UUID,
) *FunctionExecuteHeartBeat {
	return &FunctionExecuteHeartBeat{
		ID:                  value_object.NewUUID(),
		FunctionRunRecordID: functionRunRecordID,
		StartTime:           time.Now(),
		LatestHeartbeatTime: time.Now(),
	}
}

func (beb *FunctionExecuteHeartBeat) IsZero() bool {
	if beb == nil {
		return true
	}
	return beb.ID.IsNil()
}

func (beb *FunctionExecuteHeartBeat) IsTimeout(thresholdInSecond float64) bool {
	gap := time.Since(beb.LatestHeartbeatTime)
	return gap.Seconds() > thresholdInSecond
}
