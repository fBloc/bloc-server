package function_execute_heartbeat

import (
	"time"

	"github.com/fBloc/bloc-server/aggregate"
	"github.com/fBloc/bloc-server/value_object"
)

type FunctionExecuteHeartbeatRepository interface {
	// create
	Create(f *aggregate.FunctionExecuteHeartBeat) error

	// read
	GetByFunctionRunRecordID(funcRunRecordID value_object.UUID) (*aggregate.FunctionExecuteHeartBeat, error)
	AllDeads(timeoutThreshold time.Duration) ([]*aggregate.FunctionExecuteHeartBeat, error)

	// update
	AliveReportByFuncRunRecordID(functionRunRecord value_object.UUID) error

	// delete
	DeleteByFunctionRunRecordID(functionRunRecordID value_object.UUID) (int64, error)
}
