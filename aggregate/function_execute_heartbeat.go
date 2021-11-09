package aggregate

import (
	"time"

	"github.com/google/uuid"
)

type FunctionExecuteHeartBeat struct {
	ID                  uuid.UUID
	FunctionRunRecordID uuid.UUID
	StartTime           time.Time
	LatestHeartbeatTime time.Time
}

func NewFunctionExecuteHeartBeat(
	functionRunRecordID uuid.UUID,
) *FunctionExecuteHeartBeat {
	return &FunctionExecuteHeartBeat{
		ID:                  uuid.New(),
		FunctionRunRecordID: functionRunRecordID,
		StartTime:           time.Now(),
		LatestHeartbeatTime: time.Now(),
	}
}

func (beb *FunctionExecuteHeartBeat) IsZero() bool {
	if beb == nil {
		return true
	}
	return beb.ID == uuid.Nil
}

func (beb *FunctionExecuteHeartBeat) IsTimeout(thresholdInSecond float64) bool {
	gap := time.Since(beb.LatestHeartbeatTime)
	return gap.Seconds() > thresholdInSecond
}
