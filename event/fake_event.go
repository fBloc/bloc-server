package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

func init() {
	var _ DomainEvent = &FakeEvent{}
}

type FakeEvent struct {
	ID uuid.UUID
}

func (event *FakeEvent) Topic() string {
	return "fake_event"
}

// Marshal .
func (event *FakeEvent) Marshal() ([]byte, error) {
	return json.Marshal(event)
}

// Unmarshal .
func (event *FakeEvent) Unmarshal(data []byte) error {
	return json.Unmarshal(data, event)
}

// Identity
func (event *FakeEvent) Identity() string {
	return event.ID.String()
}
