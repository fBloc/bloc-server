package event

import (
	"encoding/json"

	"github.com/fBloc/bloc-server/value_object"
)

func init() {
	var _ DomainEvent = &FlowToRun{}
}

type FlowToRun struct {
	FlowRunRecordID value_object.UUID
}

func (event *FlowToRun) Topic() string {
	return "flow_task_start"
}

// Marshal .
func (event *FlowToRun) Marshal() ([]byte, error) {
	return json.Marshal(event)
}

// Unmarshal .
func (event *FlowToRun) Unmarshal(data []byte) error {
	return json.Unmarshal(data, event)
}

// Identity
func (event *FlowToRun) Identity() string {
	return event.FlowRunRecordID.String()
}
