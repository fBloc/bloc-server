package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

func init() {
	var _ DomainEvent = &FlowRunFinished{}
}

type FlowRunFinished struct {
	FlowRunRecordID uuid.UUID
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
