package event

import (
	"encoding/json"

	"github.com/fBloc/bloc-backend-go/value_object"
)

func init() {
	var _ DomainEvent = &FunctionToRun{}
}

type FunctionToRun struct {
	FunctionRunRecordID value_object.UUID
}

func (event *FunctionToRun) Topic() string {
	return "function_run_consumer"
}

// Marshal .
func (event *FunctionToRun) Marshal() ([]byte, error) {
	return json.Marshal(event)
}

// Unmarshal .
func (event *FunctionToRun) Unmarshal(data []byte) error {
	return json.Unmarshal(data, event)
}

// Identity
func (event *FunctionToRun) Identity() string {
	return event.FunctionRunRecordID.String()
}
