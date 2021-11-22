package event

import (
	"encoding/json"

	"github.com/fBloc/bloc-backend-go/value_object"
)

func init() {
	var _ DomainEvent = &FlowRunFinished{}
}

type FlowRunFinished struct {
	FlowRunRecordID value_object.UUID
}

func (event *FlowRunFinished) Topic() string {
	return "flow_task_finished"
}

// Marshal .
func (event *FlowRunFinished) Marshal() ([]byte, error) {
	return json.Marshal(event)
}

// Unmarshal .
func (event *FlowRunFinished) Unmarshal(data []byte) error {
	return json.Unmarshal(data, event)
}

// Identity
func (event *FlowRunFinished) Identity() string {
	return event.FlowRunRecordID.String()
}
