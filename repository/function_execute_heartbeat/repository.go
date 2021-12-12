package function_execute_heartbeat

import (
	"github.com/fBloc/bloc-backend-go/aggregate"
	"github.com/fBloc/bloc-backend-go/value_object"
)

type FunctionExecuteHeartbeatRepository interface {
	// create
	Create(f *aggregate.FunctionExecuteHeartBeat) error

	// read
	GetByID(id value_object.UUID) (*aggregate.FunctionExecuteHeartBeat, error)
	GetByFunctionRunRecordID(funcRunRecordID value_object.UUID) (*aggregate.FunctionExecuteHeartBeat, error)
	AllDeads() ([]*aggregate.FunctionExecuteHeartBeat, error)

	// update
	AliveReport(id value_object.UUID) error

	// delete
	Delete(id value_object.UUID) (int64, error)
}
