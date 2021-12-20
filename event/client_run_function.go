package event

import (
	"encoding/json"

	"github.com/fBloc/bloc-server/value_object"
)

func init() {
	var _ DomainEvent = &ClientRunFunction{}
}

type ClientRunFunction struct {
	FunctionRunRecordID value_object.UUID
	ClientName          string
}

func (event *ClientRunFunction) Topic() string {
	return "function_client_run_consumer." + event.ClientName
}

// Marshal .
func (event *ClientRunFunction) Marshal() ([]byte, error) {
	return json.Marshal(event)
}

// Unmarshal .
func (event *ClientRunFunction) Unmarshal(data []byte) error {
	return json.Unmarshal(data, event)
}

// Identity
func (event *ClientRunFunction) Identity() string {
	return event.FunctionRunRecordID.String()
}
