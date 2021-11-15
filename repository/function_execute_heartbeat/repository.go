package function_execute_heartbeat

import (
	"github.com/fBloc/bloc-backend-go/aggregate"

	"github.com/google/uuid"
)

type FunctionExecuteHeartbeatRepository interface {
	// create
	Create(f *aggregate.FunctionExecuteHeartBeat) error

	// read
	GetByID(id uuid.UUID) (*aggregate.FunctionExecuteHeartBeat, error)
	AllDeads() ([]*aggregate.FunctionExecuteHeartBeat, error)

	// update
	AliveReport(id uuid.UUID) error

	// delete
	Delete(id uuid.UUID) (int64, error)
}
