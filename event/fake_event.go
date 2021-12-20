package event

import (
	"encoding/json"

	"github.com/fBloc/bloc-server/value_object"
)

func init() {
	var _ DomainEvent = &FakeEvent{}
}

type FakeEvent struct {
	ID value_object.UUID
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
