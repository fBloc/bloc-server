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
	FunctionRunRecordID value_object.UUID
	LatestHeartbeatTime time.Time
}

func NewFunctionExecuteHeartBeat(
	functionRunRecordID value_object.UUID,
) *FunctionExecuteHeartBeat {
	return &FunctionExecuteHeartBeat{
		FunctionRunRecordID: functionRunRecordID,
		LatestHeartbeatTime: time.Now(),
	}
}

func (beb *FunctionExecuteHeartBeat) IsZero() bool {
	if beb == nil {
		return true
	}
	return beb.FunctionRunRecordID.IsNil()
}

func (beb *FunctionExecuteHeartBeat) IsTimeout(
	thresholdInSecond float64,
) bool {
	if beb.IsZero() {
		return false
	}
	gap := time.Since(beb.LatestHeartbeatTime)
	return gap.Seconds() > thresholdInSecond
}
